# Production-oriented Dockerfile
# Cross-compilation friendly: no QEMU needed for multi-arch builds.

# Stage 1: web build (Vite + Vue) - architecture-agnostic
FROM node:20-alpine AS webbuilder
WORKDIR /fe
COPY frontend/package*.json ./
RUN set -eux; \
  npm config set registry "https://registry.npmjs.org"; \
  npm config set fetch-retries 5; \
  npm config set fetch-retry-factor 2; \
  npm config set fetch-retry-mintimeout 20000; \
  npm config set fetch-retry-maxtimeout 120000; \
  s=1; for i in 1 2 3; do npm ci --silent --no-progress && s=0 && break || s=$?; echo "npm ci attempt $i failed: $s"; sleep 5; done; \
  if [ $s -ne 0 ]; then for i in 1 2 3; do npm install --no-audit --no-fund --legacy-peer-deps --no-progress && s=0 && break || s=$?; echo "npm install attempt $i failed: $s"; sleep 5; done; fi; \
  test $s -eq 0
COPY frontend/ ./
RUN npm run build

# Stage 2: go build (Alpine) - cross-compilation
FROM golang:1.25-alpine AS gobuilder
RUN set -eux; \
    alpine_ver=$(grep '^VERSION_ID' /etc/os-release | cut -d= -f2 | cut -d. -f1,2); \
    echo "https://dl-cdn.alpinelinux.org/alpine/v${alpine_ver}/main" > /etc/apk/repositories; \
    echo "https://dl-cdn.alpinelinux.org/alpine/v${alpine_ver}/community" >> /etc/apk/repositories; \
    for i in 1 2 3; do apk update && apk add --no-cache build-base git sqlite-dev ca-certificates tzdata wget curl unzip && break || (echo "apk failed, retry $i" && sleep 5); done
WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct \
    GOSUMDB=off \
    GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download || (go env -w GOPROXY=https://goproxy.io,direct && go mod download)
COPY . .
ENV CGO_ENABLED=1 GOOS=linux
ARG TARGETARCH
ENV GOARCH=${TARGETARCH}
ARG GIT_HASH=unknown
ARG RCLONE_VERSION=v1.73.4
RUN set -eux; \
    case "${TARGETARCH}" in \
      amd64) rclone_arch=amd64 ;; \
      arm64) rclone_arch=arm64 ;; \
      arm) rclone_arch=arm ;; \
      *) rclone_arch=amd64 ;; \
    esac; \
    ver="${RCLONE_VERSION:-v1.73.4}"; \
    base="https://github.com/rclone/rclone/releases/download/${ver}"; \
    rm -rf /tmp/rclone-extract && mkdir -p /tmp/rclone-extract /out; \
    sha_url="${base}/rclone-${ver}-linux-${rclone_arch}.zip.sha256sum"; \
    zip_url="${base}/rclone-${ver}-linux-${rclone_arch}.zip"; \
    echo "Downloading rclone ${ver} for ${rclone_arch}"; \
    curl -fsSL --retry 8 --retry-delay 2 --retry-all-errors --connect-timeout 5 \
      -o /tmp/rclone.sha256 "${sha_url}" || true; \
    if [ -s /tmp/rclone.sha256 ]; then \
      curl -fsSL --retry 8 --retry-delay 2 --retry-all-errors --connect-timeout 5 \
        -o /tmp/rclone.zip "${zip_url}"; \
      echo "Verifying rclone checksum..."; \
      cd /tmp && sha256sum -c rclone.sha256; \
      unzip -q /tmp/rclone.zip -d /tmp/rclone-extract; \
      cp /tmp/rclone-extract/rclone-*/rclone /out/rclone; \
      chmod +x /out/rclone; \
      echo "rclone verified and extracted"; \
    elif [ -s /tmp/rclone.zip ] || curl -fsSL --retry 8 --retry-delay 2 --retry-all-errors --connect-timeout 5 \
      -o /tmp/rclone.zip "${zip_url}"; then \
      echo "WARNING: SHA256SUM not available, skipping verification"; \
      unzip -q /tmp/rclone.zip -d /tmp/rclone-extract; \
      cp /tmp/rclone-extract/rclone-*/rclone /out/rclone; \
      chmod +x /out/rclone; \
    else \
      echo "rclone zip unavailable, falling back to apk rclone"; \
      (apk add --no-cache rclone || (apk update && apk add --no-cache rclone)); \
      cp /usr/bin/rclone /out/rclone; \
      chmod +x /out/rclone; \
    fi; \
    rm -rf /tmp/rclone.zip /tmp/rclone.sha256 /tmp/rclone-extract
RUN go build -ldflags="-X rcloneflow/internal/version.CommitHash=${GIT_HASH} -s -w" -o /out/server ./cmd/server

# Stage 3: runtime (Alpine)
FROM alpine:3.19
RUN set -eux; \
 apk add --no-cache bash busybox ca-certificates tzdata wget curl sqlite-libs libidn2 pcre2 su-exec; \
 mkdir -p /app/data /app/web /etc/ssl/certs /usr/share/zoneinfo
WORKDIR /app

COPY --from=gobuilder /out/server /app/server
COPY --from=webbuilder /web /app/web
COPY --from=gobuilder /out/rclone /usr/bin/rclone

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/app/server"]

EXPOSE 17870
ENV APP_ADDR=:17870
ENV APP_DATA_DIR=/app/data
ENV RCLONE_CONFIG=/app/data/rclone.conf

HEALTHCHECK NONE
