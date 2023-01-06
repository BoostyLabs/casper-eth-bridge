FROM golang:1.18 as builder
ENV GOPRIVATE github.com/BoostyLabs/golden-gate-communication
WORKDIR /app
COPY . .
RUN apt-get install git && \
    git config --global url.ssh://git@github.com/.insteadOf https://github.com/ && \
    mkdir /root/.ssh && ssh-keyscan github.com >> /root/.ssh/known_hosts
RUN --mount=type=ssh go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/polygon/main.go

# Result image
FROM alpine:3.15.4
ARG APP_DATA_DIR=/app/data
RUN mkdir -p ${APP_DATA_DIR}
COPY --from=builder /app/main .
VOLUME ["${APP_DATA_DIR}"]
