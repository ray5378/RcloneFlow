FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY web ./web
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/rclone-remote ./cmd/server

FROM alpine:3.19
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/rclone-remote /app/rclone-remote
COPY web /app/web
EXPOSE 17870
ENV APP_ADDR=:17870
ENV APP_DATA_DIR=/app/data
CMD ["/app/rclone-remote"]
