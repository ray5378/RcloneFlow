package store

import (
	"encoding/json"
	"time"
)

type User struct {
	ID              int64     `json:"id"`
	Username        string    `json:"username"`
	Password        string    `json:"-"`
	PasswordChanged bool      `json:"passwordChanged"`
	CreatedAt       time.Time `json:"createdAt"`
}

type Tag struct {
	ID   int64  `json:"id"`
	Tag  string `json:"tag"`
	Type string `json:"type"`
}

type Task struct {
	ID           int64           `json:"id"`
	Name         string          `json:"name"`
	Mode         string          `json:"mode"`
	SourceRemote string          `json:"sourceRemote"`
	SourcePath   string          `json:"sourcePath"`
	TargetRemote string          `json:"targetRemote"`
	TargetPath   string          `json:"targetPath"`
	Options      json.RawMessage `json:"options,omitempty"`
	SortOrder    int64           `json:"sortOrder"`
	CreatedAt    time.Time       `json:"createdAt"`
}

type Schedule struct {
	ID          int64      `json:"id"`
	TaskID      int64      `json:"taskId"`
	Spec        string     `json:"spec"`
	Enabled     bool       `json:"enabled"`
	NextRunTime *time.Time `json:"nextRunTime,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type Run struct {
	ID        int64          `json:"id"`
	TaskID    int64          `json:"taskId"`
	Status    string         `json:"status"`
	Trigger   string         `json:"trigger"`
	Summary   map[string]any `json:"summary,omitempty"`
	Error     string         `json:"error,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	// 任务详情
	TaskName     string `json:"taskName,omitempty"`
	TaskMode     string `json:"taskMode,omitempty"`
	SourceRemote string `json:"sourceRemote,omitempty"`
	SourcePath   string `json:"sourcePath,omitempty"`
	TargetRemote string `json:"targetRemote,omitempty"`
	TargetPath   string `json:"targetPath,omitempty"`
	// 传输详情
	FinishedAt       *time.Time `json:"finishedAt,omitempty"`
	BytesTransferred int64      `json:"bytesTransferred,omitempty"`
	Speed            string     `json:"speed,omitempty"`
}