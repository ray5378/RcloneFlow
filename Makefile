SHELL := /usr/bin/env bash

.PHONY: setup run build clean

setup:
	bash scripts/dev-setup.sh

run:
	RCLONE_BIN="$(PWD)/bin/rclone" \
	RCLONE_CONFIG="$(PWD)/data/rclone.conf" \
	go run ./cmd/server

build:
	go build -o server ./cmd/server

clean:
	rm -rf bin server
