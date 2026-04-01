package dao

import (
	"database/sql"
	"encoding/json"
	"time"

	"rcloneflow/internal/store"
)

// TaskDAO 任务数据访问对象
type TaskDAO struct {
	db *sql.DB
}

// NewTaskDAO 创建TaskDAO
func NewTaskDAO(db *sql.DB) *TaskDAO {
	return &TaskDAO{db: db}
}

// Create 创建任务
func (d *TaskDAO) Create(task store.Task) (store.Task, error) {
	result, err := d.db.Exec(`
		INSERT INTO tasks (name, mode, source_remote, source_path, target_remote, target_path)
		VALUES (?, ?, ?, ?, ?, ?)`,
		task.Name, task.Mode, task.SourceRemote, task.SourcePath, task.TargetRemote, task.TargetPath)
	if err != nil {
		return store.Task{}, err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return store.Task{}, err
	}
	
	return d.GetByID(id)
}

// GetByID 根据ID获取任务
func (d *TaskDAO) GetByID(id int64) (store.Task, bool) {
	var t store.Task
	var sourcePath, targetPath string
	err := d.db.QueryRow(`
		SELECT id, name, mode, source_remote, source_path, target_remote, target_path, created_at 
		FROM tasks WHERE id = ?`, id).Scan(
		&t.ID, &t.Name, &t.Mode, &t.SourceRemote, &sourcePath, &t.TargetRemote, &targetPath, &t.CreatedAt)
	if err != nil {
		return store.Task{}, false
	}
	t.SourcePath = sourcePath
	t.TargetPath = targetPath
	return t, true
}

// GetAll 获取所有任务
func (d *TaskDAO) GetAll() ([]store.Task, error) {
	rows, err := d.db.Query(`
		SELECT id, name, mode, source_remote, source_path, target_remote, target_path, created_at 
		FROM tasks ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var tasks []store.Task
	for rows.Next() {
		var t store.Task
		var sourcePath, targetPath string
		if err := rows.Scan(&t.ID, &t.Name, &t.Mode, &t.SourceRemote, &sourcePath, &t.TargetRemote, &targetPath, &t.CreatedAt); err != nil {
			continue
		}
		t.SourcePath = sourcePath
		t.TargetPath = targetPath
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// Update 更新任务
func (d *TaskDAO) Update(id int64, task store.Task) error {
	_, err := d.db.Exec(`
		UPDATE tasks SET name=?, mode=?, source_remote=?, source_path=?, target_remote=?, target_path=?
		WHERE id=?`,
		task.Name, task.Mode, task.SourceRemote, task.SourcePath, task.TargetRemote, task.TargetPath, id)
	return err
}

// Delete 删除任务
func (d *TaskDAO) Delete(id int64) error {
	_, err := d.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}
