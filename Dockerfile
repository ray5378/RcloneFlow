FROM debian:bookworm-slim
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata && rm -rf /var/lib/apt/lists/* && \
    useradd -m -u 1000 appuser

COPY --chown=appuser:appuser server /app/server
COPY --chown=appuser:appuser web /app/web

USER appuser

EXPOSE 17870
ENV APP_ADDR=:17870
ENV APP_DATA_DIR=/app/data

CMD ["/app/server"]
