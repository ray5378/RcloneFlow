# Stage 1: web build (Vite + Vue)
FROM node:18-alpine AS webbuilder
WORKDIR /fe
COPY frontend/ ./
RUN npm ci --silent || npm install
RUN npm run build

# Stage 2: go build (Alpine)
FROM golang:1.25-alpine AS gobuilder
# resilient apk with multi mirrors + retries to avoid TLS/permission glitches
RUN set -eux; \
    echo "https://dl-cdn.alpinelinux.org/alpine/v3.19/main" > /etc/apk/repositories; \
    echo "https://dl-cdn.alpinelinux.org/alpine/v3.19/community" >> /etc/apk/repositories; \
    echo "https://mirrors.aliyun.com/alpine/v3.19/main" >> /etc/apk/repositories; \
    echo "https://mirrors.aliyun.com/alpine/v3.19/community" >> /etc/apk/repositories; \
    for i in 1 2 3; do apk update && apk add --no-cache build-base git sqlite-dev ca-certificates tzdata wget && break || (echo "apk failed, retry $i" && sleep 5); done
WORKDIR /app
COPY go.mod ./
RUN go mod download || true
COPY . .
ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64 GOTOOLCHAIN=auto
# rclone v1.73.4
ARG RCLONE_VERSION=v1.73.4
# fallback to downloads.rclone.org to avoid Go proxy/TLS issues
RUN set -eux; (apk add --no-cache curl unzip || (apk update && apk add --no-cache curl unzip)); \
    arch="amd64"; \
    url="https://downloads.rclone.org/v1.73.4/rclone-v1.73.4-linux-${arch}.zip"; \
    curl -fSL --retry 5 --retry-connrefused -o /tmp/rclone.zip "$url"; \
    rm -rf /tmp/rclone-extract && mkdir -p /tmp/rclone-extract /out; \
    unzip -q /tmp/rclone.zip -d /tmp/rclone-extract; \
    cp /tmp/rclone-extract/rclone-*/rclone /out/rclone; \
    chmod +x /out/rclone; \
    rm -rf /tmp/rclone.zip /tmp/rclone-extract
RUN go build -o /out/server ./cmd/server

# Stage 3: runtime (Alpine)
FROM alpine:3.19
# 避免运行期访问外网：不再 apk add，改为从 gobuilder 复制所需文件
RUN adduser -D -u 1000 appuser \
 && mkdir -p /app/data /app/web /etc/ssl/certs /usr/share/zoneinfo \
 && chown -R appuser:appuser /app
WORKDIR /app

# Copy runtime deps from builder (certs, tz, wget, sqlite libs)
COPY --from=gobuilder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=gobuilder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=gobuilder /usr/bin/wget /usr/bin/wget
# sqlite shared libs
COPY --from=gobuilder /usr/lib/libsqlite3.so* /usr/lib/

# Copy server, web, and rclone
COPY --from=gobuilder /out/server /app/server
COPY --from=webbuilder /web /app/web
COPY --from=gobuilder /out/rclone /usr/bin/rclone

USER appuser

EXPOSE 17870
ENV APP_ADDR=:17870
ENV APP_DATA_DIR=/app/data
ENV RCLONE_CONFIG=/app/data/rclone.conf

# Built-in healthcheck using wget (use homepage, more robust than /healthz)
HEALTHCHECK --interval=30s --timeout=5s --start-period=45s --retries=3 \
  CMD wget -qO- http://127.0.0.1:17870/ >/dev/null 2>&1 || exit 1

CMD ["/app/server"]
