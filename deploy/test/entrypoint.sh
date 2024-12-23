#!/bin/sh
exec /app/{{service-name}}/bin/livekit-server --config=/app/{{service-name}}/conf/config.yaml --bind 0.0.0.0