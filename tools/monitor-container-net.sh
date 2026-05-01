#!/usr/bin/env bash
set -euo pipefail

CONTAINER="${1:-rclone}"
INTERVAL="${2:-2}"
COUNT="${3:-0}" # 0 = infinite

get_bytes() {
  docker exec "$CONTAINER" sh -c "cat /proc/net/dev" | awk 'NR>2 {
    gsub(":", "", $1)
    if ($1 != "lo") {
      rx += $2
      tx += $10
    }
  }
  END { printf "%0.f %0.f\n", rx, tx }'
}

human_rate() {
  awk -v bps="$1" 'function human(x) {
    split("B/s KiB/s MiB/s GiB/s", u, " ");
    i = 1;
    while (x >= 1024 && i < 4) { x /= 1024; i++ }
    return sprintf(i == 1 ? "%.0f %s" : "%.2f %s", x, u[i]);
  }
  BEGIN { print human(bps) }'
}

if ! docker ps --format '{{.Names}}' | grep -Fxq "$CONTAINER"; then
  echo "container not running: $CONTAINER" >&2
  exit 1
fi

read -r prev_rx prev_tx < <(get_bytes)
prev_ts=$(date +%s)
iter=0

echo "Monitoring container=$CONTAINER interval=${INTERVAL}s (Ctrl+C to stop)"
printf '%-8s %-12s %-12s %-12s %-14s %-14s\n' "TIME" "RX" "TX" "TOTAL" "RX_TOTAL" "TX_TOTAL"

while true; do
  sleep "$INTERVAL"
  read -r rx tx < <(get_bytes)
  now=$(date +%s)
  dt=$(( now - prev_ts ))
  if [ "$dt" -le 0 ]; then
    dt=1
  fi
  drx=$(( rx - prev_rx ))
  dtx=$(( tx - prev_tx ))
  if [ "$drx" -lt 0 ]; then drx=0; fi
  if [ "$dtx" -lt 0 ]; then dtx=0; fi
  total=$(( drx + dtx ))

  rx_rate=$(human_rate $(( drx / dt )))
  tx_rate=$(human_rate $(( dtx / dt )))
  total_rate=$(human_rate $(( total / dt )))
  rx_total=$(human_rate "$rx")
  tx_total=$(human_rate "$tx")

  printf '%-8s %-12s %-12s %-12s %-14s %-14s\n' "$(date +%H:%M:%S)" "$rx_rate" "$tx_rate" "$total_rate" "$rx_total" "$tx_total"

  prev_rx=$rx
  prev_tx=$tx
  prev_ts=$now
  iter=$(( iter + 1 ))
  if [ "$COUNT" -gt 0 ] && [ "$iter" -ge "$COUNT" ]; then
    break
  fi
done
