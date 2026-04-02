FROM debian:bookworm-slim
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

COPY server /app/rcloneflow
COPY web /app/web

EXPOSE 17870
ENV APP_ADDR=:17870
ENV APP_DATA_DIR=/app/data

CMD ["/app/rcloneflow"]
