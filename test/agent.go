// Copyright 2024 LiveKit, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/p1cn/livekit-protocol-extension/livekitext"
	"go.uber.org/atomic"
	"google.golang.org/protobuf/proto"

	"github.com/livekit/protocol/livekit"
)

type agentClient struct {
	mu   sync.Mutex
	conn *websocket.Conn

	registered              atomic.Int32
	roomAvailability        atomic.Int32
	roomJobs                atomic.Int32
	participantAvailability atomic.Int32
	participantJobs         atomic.Int32

	requestedJobs chan *livekit.Job

	done chan struct{}
}

func newAgentClient(token string, port uint32) (*agentClient, error) {
	host := fmt.Sprintf("ws://localhost:%d", port)
	u, err := url.Parse(host + "/agent")
	if err != nil {
		return nil, err
	}
	requestHeader := make(http.Header)
	requestHeader.Set("Authorization", "Bearer "+token)

	connectUrl := u.String()
	conn, _, err := websocket.DefaultDialer.Dial(connectUrl, requestHeader)
	if err != nil {
		return nil, err
	}

	return &agentClient{
		conn:          conn,
		requestedJobs: make(chan *livekit.Job, 100),
		done:          make(chan struct{}),
	}, nil
}

func (c *agentClient) Run(jobType livekit.JobType, namespace string) (err error) {
	go c.read()

	switch jobType {
	case livekit.JobType_JT_ROOM:
		err = c.write(&livekitext.WorkerMessage{
			Message: &livekitext.WorkerMessage_Register{
				Register: &livekit.RegisterWorkerRequest{
					Type:      livekit.JobType_JT_ROOM,
					Version:   "version",
					Namespace: &namespace,
				},
			},
		})

	case livekit.JobType_JT_PUBLISHER:
		err = c.write(&livekitext.WorkerMessage{
			Message: &livekitext.WorkerMessage_Register{
				Register: &livekit.RegisterWorkerRequest{
					Type:      livekit.JobType_JT_PUBLISHER,
					Version:   "version",
					Namespace: &namespace,
				},
			},
		})
	}

	return err
}

func (c *agentClient) read() {
	for {
		select {
		case <-c.done:
			return
		default:
			_, b, err := c.conn.ReadMessage()
			if err != nil {
				return
			}

			msg := &livekit.ServerMessage{}
			if err = proto.Unmarshal(b, msg); err != nil {
				return
			}

			switch m := msg.Message.(type) {
			case *livekit.ServerMessage_Assignment:
				go c.handleAssignment(m.Assignment)
			case *livekit.ServerMessage_Availability:
				go c.handleAvailability(m.Availability)
			case *livekit.ServerMessage_Register:
				go c.handleRegister(m.Register)
			}
		}
	}
}

func (c *agentClient) handleAssignment(req *livekit.JobAssignment) {
	if req.Job.Type == livekit.JobType_JT_ROOM {
		c.roomJobs.Inc()
	} else {
		c.participantJobs.Inc()
	}
}

func (c *agentClient) handleAvailability(req *livekit.AvailabilityRequest) {
	if req.Job.Type == livekit.JobType_JT_ROOM {
		c.roomAvailability.Inc()
	} else {
		c.participantAvailability.Inc()
	}

	c.requestedJobs <- req.Job

	c.write(&livekitext.WorkerMessage{
		Message: &livekitext.WorkerMessage_Availability{
			Availability: &livekit.AvailabilityResponse{
				JobId:     req.Job.Id,
				Available: true,
			},
		},
	})
}

func (c *agentClient) handleRegister(req *livekit.RegisterWorkerResponse) {
	c.registered.Inc()
}

func (c *agentClient) write(msg *livekitext.WorkerMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.done:
		return nil
	default:
		b, err := proto.Marshal(msg)
		if err != nil {
			return err
		}

		return c.conn.WriteMessage(websocket.BinaryMessage, b)
	}
}

func (c *agentClient) close() {
	c.mu.Lock()
	close(c.done)
	_ = c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	_ = c.conn.Close()
	c.mu.Unlock()
}
