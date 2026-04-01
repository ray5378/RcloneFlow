FROM golang:1.22-bookworm AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download || true
COPY cmd ./cmd
COPY internal ./internal
COPY web ./web
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/rclone-remote ./cmd/server

FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates tzdata \
  && rm -rf /var/lib/apt/lists/*
COPY --from=builder /out/rclone-remote /app/rclone-remote
COPY web /app/web
EXPOSE 17870
ENV APP_ADDR=:17870
ENV APP_DATA_DIR=/app/data
CMD ["/app/rclone-remote"]
