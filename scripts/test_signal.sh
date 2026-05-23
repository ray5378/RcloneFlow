#!/bin/bash
# scripts/test_signal.sh — Docker 容器信号测试
# 覆盖 main() 和 app.Run() 中非可测的 signal.Notify 路径。
# 需要在项目根目录执行。

set -euo pipefail

IMAGE="rcloneflow-signal-test:latest"
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PASS=0
FAIL=0

ok()   { PASS=$((PASS+1)); echo "  ✓ PASS"; }
fail() { FAIL=$((FAIL+1)); echo "  ✗ FAIL"; }

OUTFILE="$(mktemp)"

cleanup() {
  if [ -n "${CID:-}" ]; then
    docker rm -f "$CID" >/dev/null 2>&1 || true
  fi
  rm -f "$OUTFILE"
}
trap cleanup EXIT

# ---------------------------------------------------------------
echo "=== 构建镜像 ==="
docker build -t "$IMAGE" "$PROJECT_ROOT"
echo ""

# ---------------------------------------------------------------
echo "=== Test 1: main() 启动失败 → stderr 输出 + os.Exit(1) ==="
set +e
docker run --rm \
  -e JWT_SECRET=test \
  -e EMBED_RC=false \
  -e APP_DATA_DIR=/dev/null/bad \
  "$IMAGE" >"$OUTFILE" 2>&1
EC=$?
set -e
grep -q "启动失败" "$OUTFILE" && ok || fail
[ "$EC" -eq 1 ] && ok || { echo "  (exit code was $EC, expected 1)"; fail; }
echo ""

# ---------------------------------------------------------------
echo "=== Test 2: app.Run() SIGINT 关闭 → signal.Notify 路径被覆盖 ==="
CID=$(docker run -d \
  -e JWT_SECRET=test \
  -e EMBED_RC=false \
  -e APP_DATA_DIR=/tmp/rf-testdata \
  -e LOG_LEVEL=info \
  "$IMAGE")
echo "  container: $CID"

for i in $(seq 1 15); do
  if docker logs "$CID" 2>/dev/null | grep -q "服务监听中"; then
    echo "  server ready after ${i}s"
    break
  fi
  sleep 1
done

docker kill --signal=SIGINT "$CID" >/dev/null
echo "  SIGINT sent"

sleep 3
LOGS=$(docker logs "$CID" 2>&1)

echo "$LOGS" | grep -q "收到关闭信号" && ok || fail
echo "$LOGS" | grep -q "服务器已安全关闭" && ok || fail
echo ""

# ---------------------------------------------------------------
echo "=== Test 3: 容器最终退出码 0（正常关闭） ==="
for i in $(seq 1 10); do
  EC=$(docker inspect "$CID" --format '{{.State.ExitCode}}' 2>/dev/null || echo "running")
  [ "$EC" != "running" ] && break
  sleep 1
done
[ "$EC" = "0" ] && ok || { echo "  (exit code was $EC, expected 0)"; fail; }
echo ""

# ---------------------------------------------------------------
echo "=== 汇总 ==="
echo "  PASS: $PASS  FAIL: $FAIL"
if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
