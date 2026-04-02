FROM golang:1.24-alpine
WORKDIR /app

RUN apk add --no-install-recommends ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY web ./web

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o rcloneflow ./cmd/server

EXPOSE 17870
ENV APP_ADDR=:17870
ENV APP_DATA_DIR=/app/data

CMD ["./rcloneflow"]
