#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CACHE_DIR="${ROOT_DIR}/third_party/docker"

IMAGES=(
  "node:18-alpine"
  "golang:1.25-alpine"
  "alpine:3.19"
)

sanitize_image_name() {
  local image="$1"
  image="${image//\//-}"
  image="${image//:/-}"
  echo "$image"
}

has_image() {
  local image="$1"
  docker image inspect "$image" >/dev/null 2>&1
}

load_from_tar_if_exists() {
  local image="$1"
  local tar_name
  tar_name="$(sanitize_image_name "$image").tar"
  local tar_path="${CACHE_DIR}/${tar_name}"

  if [[ -f "$tar_path" ]]; then
    echo "[load-base-images] loading cached tar: $tar_path"
    docker load -i "$tar_path" >/dev/null
    return 0
  fi

  return 1
}

pull_image() {
  local image="$1"
  echo "[load-base-images] pulling $image"
  docker pull "$image"
}

ensure_image() {
  local image="$1"

  if has_image "$image"; then
    echo "[load-base-images] already present: $image"
    return 0
  fi

  if load_from_tar_if_exists "$image"; then
    if has_image "$image"; then
      echo "[load-base-images] loaded from cache: $image"
      return 0
    fi
    echo "[load-base-images] warning: tar loaded but image tag not found as $image, fallback to pull"
  fi

  pull_image "$image"
}

main() {
  command -v docker >/dev/null 2>&1 || {
    echo "[load-base-images] docker not found" >&2
    exit 1
  }

  mkdir -p "$CACHE_DIR"

  for image in "${IMAGES[@]}"; do
    ensure_image "$image"
  done

  echo "[load-base-images] done"
}

main "$@"
