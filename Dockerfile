# Stage 1: fetch rclone (latest)
FROM alpine:3.19 AS rclone-fetch
ARG RCLONE_VERSION=latest
RUN apk add --no-cache ca-certificates curl unzip && \
    if [ "$RCLONE_VERSION" = "latest" ] || [ "$RCLONE_VERSION" = "current" ]; then \
      URL="https://downloads.rclone.org/rclone-current-linux-amd64.zip"; \
    else \
      URL="https://downloads.rclone.org/${RCLONE_VERSION}/rclone-${RCLONE_VERSION}-linux-amd64.zip"; \
    fi && \
    curl -fsSL "$URL" -o /tmp/rclone.zip && \
    unzip /tmp/rclone.zip -d /tmp && \
    mv /tmp/rclone-*-linux-amd64/rclone /usr/local/bin/rclone && \
    chmod +x /usr/local/bin/rclone

# Stage 2: final runtime
FROM alpine:3.19
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata wget && \
    adduser -D -u 1000 appuser

COPY --chown=appuser:appuser server /app/server
COPY --chown=appuser:appuser web /app/web
COPY --from=rclone-fetch /usr/local/bin/rclone /usr/local/bin/rclone

USER appuser

EXPOSE 17870
ENV APP_ADDR=:17870
ENV APP_DATA_DIR=/app/data

HEALTHCHECK --interval=30s --timeout=3s --retries=3 CMD wget -q -O- http://127.0.0.1:17870/healthz || exit 1

CMD ["/app/server"]
