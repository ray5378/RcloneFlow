.PHONY: build test lint clean frontend-install frontend-build frontend-test dev

COMMIT_HASH := $(shell git rev-parse --short=7 HEAD 2>/dev/null || echo unknown)
LDFLAGS := -s -w -X rcloneflow/internal/version.CommitHash=$(COMMIT_HASH)

build:
	go build -ldflags="$(LDFLAGS)" -o bin/rcloneflow-server ./cmd/server
	cd frontend && npm run build

test:
	go test -count=1 ./...

frontend-install:
	cd frontend && npm ci --legacy-peer-deps

frontend-build:
	cd frontend && npm run build

frontend-test:
	cd frontend && npm run test

lint:
	go vet ./...

dev:
	go run ./cmd/server &

clean:
	rm -rf bin/
	rm -f coverage.out