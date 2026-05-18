package service

import "errors"

var (
	// ErrTaskNotFound 任务未找到
	ErrTaskNotFound = errors.New("task not found")

	// ErrTaskNameExists 任务名已存在
	ErrTaskNameExists = errors.New("task name already exists")

	// ErrScheduleNotFound 定时任务未找到
	ErrScheduleNotFound = errors.New("schedule not found")

	// ErrRunNotFound 运行记录未找到
	ErrRunNotFound = errors.New("run not found")
)
