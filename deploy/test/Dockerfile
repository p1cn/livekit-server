# source version
ARG version
# env, dev, test, prod
ARG env
# harbor host
ARG harbor

# step1 build
FROM ${harbor}/cicd/go:1.23.3 as builder
WORKDIR /app

COPY ./go.mod /app/
COPY ./go.sum /app/
ENV GOPROXY=https://goproxy.cn,direct
RUN source /etc/profile && GOPROXY=https://goproxy.cn,direct go mod download

ARG version
ARG env
ARG service
ADD ./ /app
RUN source /etc/profile && make pack ENV=${env} VERSION=${version} SERVICE_NAME=${service}

# step2 run
FROM ${harbor}/oss/centos_sv:latest
ARG env
ARG service
VOLUME /root/go/pkg
VOLUME /root/go/src
VOLUME /root/.cache/go-build
COPY --from=builder /app/build /app/${service}/
COPY deploy/${env}/entrypoint.sh /
RUN chmod +x /entrypoint.sh
ENTRYPOINT /bin/sh -c /entrypoint.sh