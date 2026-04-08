FROM debian:bookworm-slim
WORKDIR /app

# Install dependencies and rclone binary
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates tzdata curl unzip \
 && rm -rf /var/lib/apt/lists/* \
 && useradd -m -u 1000 appuser \
 && mkdir -p /app/data \
 && chown -R appuser:appuser /app

# Fetch rclone (current linux/amd64) and install to /usr/bin/rclone
RUN curl -fsSL -o /tmp/rclone.zip https://downloads.rclone.org/rclone-current-linux-amd64.zip \
 && unzip -q /tmp/rclone.zip -d /tmp \
 && cp /tmp/rclone-*-linux-amd64/rclone /usr/bin/rclone \
 && chmod +x /usr/bin/rclone \
 && rm -rf /tmp/rclone.zip /tmp/rclone-*-linux-amd64

# Copy server and web assets
COPY --chown=appuser:appuser server /app/server
COPY --chown=appuser:appuser web /app/web

USER appuser

EXPOSE 17870
ENV APP_ADDR=:17870
ENV APP_DATA_DIR=/app/data
# Unified rclone config path shared by RC and CLI
ENV RCLONE_CONFIG=/app/data/rclone.conf

CMD ["/app/server"]
