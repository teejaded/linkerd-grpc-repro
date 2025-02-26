FROM golang:1.24 as builder

ENV GOPROXY=https://proxy.golang.org

WORKDIR /app

COPY go.mod go.sum ./

COPY keepalive/server ./keepalive/server
RUN --mount=type=cache,target="/root/.cache/go-build" \
    CGO_ENABLED=0 go build -o server ./keepalive/server

COPY keepalive/client ./keepalive/client
RUN --mount=type=cache,target="/root/.cache/go-build" \
    CGO_ENABLED=0 go build -o client ./keepalive/client

FROM scratch

COPY --from=builder /app/server /
COPY --from=builder /app/client /
