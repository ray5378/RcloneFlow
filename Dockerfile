# Stage 1: build (Alpine)
FROM golang:1.22-alpine AS builder
RUN apk add --no-cache build-base git sqlite-dev ca-certificates
WORKDIR /app
COPY go.mod ./
RUN go mod download || true
COPY . .
# go-sqlite3 需要 CGO，Alpine 使用 musl，需启用 CGO 并安装 sqlite-dev
ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64
RUN go build -o /out/server ./cmd/server

# Stage 2: runtime (Alpine)
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata sqlite-libs rclone \
 && adduser -D -u 1000 appuser \
 && mkdir -p /app/data /app/web \
 && chown -R appuser:appuser /app
WORKDIR /app

# Copy server and web assets
COPY --from=builder /out/server /app/server
COPY web /app/web

USER appuser

EXPOSE 17870
ENV APP_ADDR=:17870
ENV APP_DATA_DIR=/app/data
# 统一 rclone 配置路径（RC 与 CLI 共用）
ENV RCLONE_CONFIG=/app/data/rclone.conf

CMD ["/app/server"]
