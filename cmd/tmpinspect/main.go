package main

import (
  "fmt"
  "log"
  "rcloneflow/internal/store"
)

func main() {
  db, err := store.Open("/docker/rcloneflow/data")
  if err != nil { log.Fatal(err) }
  runs, total, err := db.ListRuns(1, 20)
  if err != nil { log.Fatal(err) }
  fmt.Printf("total=%d\n", total)
  for _, r := range runs {
    fmt.Printf("id=%d task=%d status=%q created=%s finished=%v\n", r.ID, r.TaskID, r.Status, r.CreatedAt.Format("2006-01-02 15:04:05"), r.FinishedAt)
  }
  active, err := db.ListActiveRuns()
  if err != nil { log.Fatal(err) }
  fmt.Printf("active_count=%d\n", len(active))
  for _, r := range active {
    fmt.Printf("ACTIVE id=%d task=%d status=%q\n", r.ID, r.TaskID, r.Status)
  }
}
