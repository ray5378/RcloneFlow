package runnercli

import (
	"rcloneflow/internal/store"
	"rcloneflow/internal/websocket"
)

// RunUpdater 抽象 run 更新操作，解耦 runnercli 与 store.DB
type RunUpdater interface {
	UpdateRun(id int64, fn func(*store.Run)) error
	GetRun(id int64) (store.Run, error)
	GetTask(id int64) (store.Task, bool)
}

// EventBroadcaster 抽象事件广播，解耦 runnercli 与 websocket
type EventBroadcaster interface {
	Broadcast(eventType string, data map[string]any)
}

// StoreDBAdapter 将 *store.DB 适配为 RunUpdater
type StoreDBAdapter struct {
	DB *store.DB
}

func (a *StoreDBAdapter) UpdateRun(id int64, fn func(*store.Run)) error {
	return a.DB.UpdateRun(id, fn)
}

func (a *StoreDBAdapter) GetRun(id int64) (store.Run, error) {
	return a.DB.GetRun(id)
}

func (a *StoreDBAdapter) GetTask(id int64) (store.Task, bool) {
	return a.DB.GetTask(id)
}

// WSBroadcaster 将 websocket.Broadcast 适配为 EventBroadcaster
type WSBroadcaster struct{}

func (b *WSBroadcaster) Broadcast(eventType string, data map[string]any) {
	websocket.Broadcast(eventType, data)
}
