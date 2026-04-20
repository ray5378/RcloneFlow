# Production-oriented Dockerfile
# 已移除对仓库内 third_party 构建缓存的依赖，避免收尾阶段继续耦合本地临时缓存。

# Stage 1: web build (Vite + Vue)
FROM node:18-alpine AS webbuilder
WORKDIR /fe
COPY frontend/package*.json ./
ENV NPM_CONFIG_REGISTRY=https://registry.npmmirror.com
RUN set -eux; \
  npm config set registry "$NPM_CONFIG_REGISTRY"; \
  npm config set fetch-retries 5; \
  npm config set fetch-retry-factor 2; \
  npm config set fetch-retry-mintimeout 20000; \
  npm config set fetch-retry-maxtimeout 120000; \
  s=1; for i in 1 2 3; do npm ci --silent --no-progress && s=0 && break || s=$?; echo "npm ci attempt $i failed: $s"; sleep 5; done; \
  if [ $s -ne 0 ]; then for i in 1 2 3; do npm install --no-audit --no-fund --legacy-peer-deps --no-progress && s=0 && break || s=$?; echo "npm install attempt $i failed: $s"; sleep 5; done; fi; \
  test $s -eq 0
COPY frontend/ ./
RUN npm run build

# Stage 2: go build (Alpine)
FROM golang:1.25-alpine AS gobuilder
RUN set -eux; \
    echo "https://dl-cdn.alpinelinux.org/alpine/v3.19/main" > /etc/apk/repositories; \
    echo "https://dl-cdn.alpinelinux.org/alpine/v3.19/community" >> /etc/apk/repositories; \
    echo "https://mirrors.aliyun.com/alpine/v3.19/main" >> /etc/apk/repositories; \
    echo "https://mirrors.aliyun.com/alpine/v3.19/community" >> /etc/apk/repositories; \
    for i in 1 2 3; do apk update && apk add --no-cache build-base git sqlite-dev ca-certificates tzdata wget curl unzip && break || (echo "apk failed, retry $i" && sleep 5); done
WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct \
    GOSUMDB=off \
    GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download || (go env -w GOPROXY=https://goproxy.io,direct && go mod download)
COPY . .
ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64
ARG RCLONE_VERSION=v1.73.4
RUN set -eux; \
    arch="$(apk --print-arch)"; \
    case "$arch" in \
      x86_64) arch=amd64 ;; \
      aarch64) arch=arm64 ;; \
      armhf) arch=arm ;; \
      *) arch=amd64 ;; \
    esac; \
    ver="${RCLONE_VERSION:-v1.73.4}"; \
    urls="https://github.com/rclone/rclone/releases/download/${ver}/rclone-${ver}-linux-${arch}.zip https://downloads.rclone.org/${ver}/rclone-${ver}-linux-${arch}.zip"; \
    rm -f /tmp/rclone.zip; \
    for u in $urls; do \
      echo "Trying $u"; \
      if curl -fsSL --retry 8 --retry-delay 2 --retry-all-errors --connect-timeout 5 -o /tmp/rclone.zip "$u"; then \
        break; \
      fi; \
    done; \
    rm -rf /tmp/rclone-extract && mkdir -p /tmp/rclone-extract /out; \
    if [ -s /tmp/rclone.zip ]; then \
      unzip -q /tmp/rclone.zip -d /tmp/rclone-extract; \
      cp /tmp/rclone-extract/rclone-*/rclone /out/rclone; \
      chmod +x /out/rclone; \
      rm -rf /tmp/rclone.zip /tmp/rclone-extract; \
    else \
      echo "rclone zip unavailable, falling back to apk rclone"; \
      (apk add --no-cache rclone || (apk update && apk add --no-cache rclone)); \
      cp /usr/bin/rclone /out/rclone; \
      chmod +x /out/rclone; \
    fi
RUN go build -ldflags="-s -w" -o /out/server ./cmd/server

# Stage 3: runtime (Alpine)
FROM alpine:3.19
RUN adduser -D -u 1000 appuser \
 && mkdir -p /app/data /app/web /etc/ssl/certs /usr/share/zoneinfo \
 && chown -R appuser:appuser /app
WORKDIR /app

COPY --from=gobuilder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=gobuilder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=gobuilder /usr/bin/wget /usr/bin/wget
COPY --from=gobuilder /usr/lib/libsqlite3.so* /usr/lib/
COPY --from=gobuilder /out/server /app/server
COPY --from=webbuilder /web /app/web
COPY --from=gobuilder /out/rclone /usr/bin/rclone

USER appuser

EXPOSE 17870
ENV APP_ADDR=:17870
ENV APP_DATA_DIR=/app/data
ENV RCLONE_CONFIG=/app/data/rclone.conf

HEALTHCHECK NONE

CMD ["/app/server"]
