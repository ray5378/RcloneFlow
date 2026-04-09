# Stage 1: web build (Vite + Vue)
FROM node:18-alpine AS webbuilder
WORKDIR /fe
COPY frontend/ ./
RUN npm ci --silent || npm install
RUN npm run build

# Stage 2: go build (Alpine)
FROM golang:1.25-alpine AS gobuilder
RUN apk add --no-cache build-base git sqlite-dev ca-certificates
WORKDIR /app
COPY go.mod ./
RUN go mod download || true
COPY . .
ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64 GOTOOLCHAIN=auto
# rclone v1.73.4
ARG RCLONE_VERSION=v1.73.4
# fallback to downloads.rclone.org to avoid Go proxy/TLS issues
RUN set -eux; apk add --no-cache curl unzip; \
    arch="amd64"; \
    url="https://downloads.rclone.org/v1.73.4/rclone-v1.73.4-linux-${arch}.zip"; \
    curl -fSL --retry 5 --retry-connrefused -o /tmp/rclone.zip "$url"; \
    rm -rf /tmp/rclone-extract && mkdir -p /tmp/rclone-extract; \
    unzip -q /tmp/rclone.zip -d /tmp/rclone-extract; \
    cp /tmp/rclone-extract/rclone-*/rclone /out/rclone; \
    chmod +x /out/rclone; \
    rm -rf /tmp/rclone.zip /tmp/rclone-extract
RUN go build -o /out/server ./cmd/server

# Stage 3: runtime (Alpine)
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata sqlite-libs \
 && adduser -D -u 1000 appuser \
 && mkdir -p /app/data /app/web \
 && chown -R appuser:appuser /app
WORKDIR /app

# Copy server, web, and rclone
COPY --from=gobuilder /out/server /app/server
COPY --from=webbuilder /web /app/web
COPY --from=gobuilder /out/rclone /usr/bin/rclone

USER appuser

EXPOSE 17870
ENV APP_ADDR=:17870
ENV APP_DATA_DIR=/app/data
ENV RCLONE_CONFIG=/app/data/rclone.conf

CMD ["/app/server"]
