FROM golang:1.18 as builder
WORKDIR /app
COPY . .
RUN apt-get install git && \
    git config --global url.ssh://git@github.com/.insteadOf https://github.com/ && \
    mkdir /root/.ssh && ssh-keyscan github.com >> /root/.ssh/known_hosts
RUN --mount=type=ssh go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/gateway/main.go

# Result image
FROM alpine:3.15.4
COPY --from=builder /app/main .
COPY --from=builder /app/bridge/gateway/docs/console ./bridge/gateway/docs/console
