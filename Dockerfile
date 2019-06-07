#FROM golang:1.10
FROM golang:1.11.5-alpine as builder
RUN /sbin/apk update && /sbin/apk add --no-cache git gcc bind-dev musl-dev
ADD . /app/
WORKDIR /app/
RUN go mod download
RUN go build -o /app/app .
#WORKDIR /app/plugins/timelog/ 
#RUN go build -buildmode=plugin -o /app/plugins/timelog/timelog-plugin.so 
run find ./ -name "*.go" -delete
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/plugins ./app/plugins
COPY --from=builder /app/app ./app/app
COPY --from=builder /app/pool.config.json ./app/pool.config.json
COPY --from=builder /app/server.crt ./app/server.crt
COPY --from=builder /app/server.key ./app/server.key
COPY --from=builder /app/pool.config.json ./app/pool.config.json
RUN echo "net.ipv4.tcp_tw_reuse = 1" >> /etc/sysctl.conf
RUN echo "net.ipv4.tcp_tw_recycle = 1" >> /etc/sysctl.conf
RUN echo "net.ipv4.ip_local_port_range = 50000 " >> /etc/sysctl.conf


WORKDIR /root/app/
run ls
ENTRYPOINT ["./app", "pool.config.json"]
