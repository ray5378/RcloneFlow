# TEMP TEST-ACCEL VARIANT
# 说明：这是为了“快速测试、尽量减少联网失败”而加的临时方案。
# 优先使用仓库内缓存（基础镜像预加载、npm cache、go mod cache、apk cache），
# 后续开发稳定后应回收/剔除，不作为长期规范默认方案。

# Stage 1: web build (Vite + Vue)
FROM node:18-alpine AS webbuilder
WORKDIR /fe
COPY frontend/package*.json ./
COPY third_party/npm-cache /npm-cache
# TEMP: prefer repo-local npm cache first; fallback to mirror/network when cache misses
ENV NPM_CONFIG_REGISTRY=https://registry.npmmirror.com \
    NPM_CONFIG_CACHE=/npm-cache
RUN set -eux; \
  npm config set registry "$NPM_CONFIG_REGISTRY"; \
  npm config set cache "$NPM_CONFIG_CACHE"; \
  npm config set fetch-retries 5; \
  npm config set fetch-retry-factor 2; \
  npm config set fetch-retry-mintimeout 20000; \
  npm config set fetch-retry-maxtimeout 120000; \
  s=1; for i in 1 2 3; do npm ci --cache "$NPM_CONFIG_CACHE" --prefer-offline --silent --no-progress && s=0 && break || s=$?; echo "npm ci attempt $i failed: $s"; sleep 5; done; \
  if [ $s -ne 0 ]; then for i in 1 2 3; do npm install --cache "$NPM_CONFIG_CACHE" --prefer-offline --no-audit --no-fund --legacy-peer-deps --no-progress && s=0 && break || s=$?; echo "npm install attempt $i failed: $s"; sleep 5; done; fi; \
  test $s -eq 0
COPY frontend/ ./
RUN npm run build

# Stage 2: go build (Alpine)
FROM golang:1.25-alpine AS gobuilder
COPY third_party/apk-cache /apk-cache
# TEMP: prefer repo-local apk cache first; fallback to network mirrors when cache misses
RUN set -eux; \
    if ls /apk-cache/*.apk >/dev/null 2>&1; then \
      apk add --no-network --allow-untrusted /apk-cache/*.apk || true; \
    fi; \
    if ! apk info -e build-base git sqlite-dev ca-certificates tzdata wget >/dev/null 2>&1; then \
      echo "https://dl-cdn.alpinelinux.org/alpine/v3.19/main" > /etc/apk/repositories; \
      echo "https://dl-cdn.alpinelinux.org/alpine/v3.19/community" >> /etc/apk/repositories; \
      echo "https://mirrors.aliyun.com/alpine/v3.19/main" >> /etc/apk/repositories; \
      echo "https://mirrors.aliyun.com/alpine/v3.19/community" >> /etc/apk/repositories; \
      for i in 1 2 3; do apk update && apk add --no-cache build-base git sqlite-dev ca-certificates tzdata wget && break || (echo "apk failed, retry $i" && sleep 5); done; \
    fi
WORKDIR /app
# TEMP: prefer repo-local Go module cache first; fallback to public proxy when cache misses
ENV GOPROXY=file:///go-mod-cache/cache/download,https://goproxy.cn,direct \
    GOSUMDB=off \
    GOMODCACHE=/go-mod-cache
COPY go.mod go.sum ./
COPY third_party/go-mod-cache /go-mod-cache
RUN go mod download || (go env -w GOPROXY=https://goproxy.io,direct && go mod download) || true
COPY . .
ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64 GOTOOLCHAIN=auto
# rclone v1.73.4
ARG RCLONE_VERSION=v1.73.4
# Prefer repo-local cached rclone zip first; if absent, download remotely; final fallback to apk rclone
RUN set -eux; \
    if ls /apk-cache/*.apk >/dev/null 2>&1; then \
      apk add --no-network --allow-untrusted /apk-cache/*.apk || true; \
    fi; \
    if ! apk info -e curl unzip >/dev/null 2>&1; then \
      (apk add --no-cache curl unzip || (apk update && apk add --no-cache curl unzip)); \
    fi; \
    arch="$(apk --print-arch)"; \
    case "$arch" in \
      x86_64) arch=amd64 ;; \
      aarch64) arch=arm64 ;; \
      armhf) arch=arm ;; \
      *) arch=amd64 ;; \
    esac; \
    ver="${RCLONE_VERSION:-v1.73.4}"; \
    local_zip="/app/third_party/rclone/rclone-${ver}-linux-${arch}.zip"; \
    urls="https://github.com/rclone/rclone/releases/download/${ver}/rclone-${ver}-linux-${arch}.zip https://downloads.rclone.org/${ver}/rclone-${ver}-linux-${arch}.zip"; \
    rm -f /tmp/rclone.zip; \
    if [ -s "$local_zip" ]; then \
      echo "Using local cached rclone zip: $local_zip"; \
      cp "$local_zip" /tmp/rclone.zip; \
    else \
      for u in $urls; do \
        echo "Trying $u"; \
        if curl -fsSL --retry 8 --retry-delay 2 --retry-all-errors --connect-timeout 5 -o /tmp/rclone.zip "$u"; then \
          break; \
        fi; \
      done; \
    fi; \
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
# 避免运行期访问外网：不再 apk add，改为从 gobuilder 复制所需文件
RUN adduser -D -u 1000 appuser \
 && mkdir -p /app/data /app/web /etc/ssl/certs /usr/share/zoneinfo \
 && chown -R appuser:appuser /app
WORKDIR /app

# Copy runtime deps from builder (certs, tz, sqlite libs, wget)
COPY --from=gobuilder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=gobuilder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=gobuilder /usr/bin/wget /usr/bin/wget
# sqlite shared libs
COPY --from=gobuilder /usr/lib/libsqlite3.so* /usr/lib/

# Copy server, web, and rclone
COPY --from=gobuilder /out/server /app/server
# Copy built frontend (Vite outDir '../web') into runtime web dir
COPY --from=webbuilder /web /app/web
COPY --from=gobuilder /out/rclone /usr/bin/rclone

USER appuser

EXPOSE 17870
ENV APP_ADDR=:17870
ENV APP_DATA_DIR=/app/data
ENV RCLONE_CONFIG=/app/data/rclone.conf

# No built-in healthcheck; external systems can probe /healthz
HEALTHCHECK NONE

CMD ["/app/server"]
