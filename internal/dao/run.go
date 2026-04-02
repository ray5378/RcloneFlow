package dao

import (
	"database/sql"
	"encoding/json"
	"time"

	"rcloneflow/internal/store"
)

// RunDAO 运行记录数据访问对象
type RunDAO struct {
	db *sql.DB
}

// NewRunDAO 创建RunDAO
func NewRunDAO(db *sql.DB) *RunDAO {
	return &RunDAO{db: db}
}

// Create 创建运行记录
func (d *RunDAO) Create(run store.Run) (store.Run, error) {
	result, err := d.db.Exec(`
		INSERT INTO runs (task_id, rc_job_id, status, trigger)
		VALUES (?, ?, ?, ?)`,
		run.TaskID, run.RcJobID, run.Status, run.Trigger)
	if err != nil {
		return store.Run{}, err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return store.Run{}, err
	}
	
	return d.GetByID(id)
}

// GetByID 根据ID获取运行记录
func (d *RunDAO) GetByID(id int64) (store.Run, error) {
	var r store.Run
	var summaryJSON string
	err := d.db.QueryRow(`
		SELECT id, task_id, rc_job_id, status, trigger, summary, error, created_at, updated_at
		FROM runs WHERE id = ?`, id).Scan(
		&r.ID, &r.TaskID, &r.RcJobID, &r.Status, &r.Trigger, &summaryJSON, &r.Error, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return store.Run{}, err
	}
	if summaryJSON != "" && summaryJSON != "{}" {
		json.Unmarshal([]byte(summaryJSON), &r.Summary)
	}
	return r, nil
}

// GetAll 获取所有运行记录
func (d *RunDAO) GetAll() ([]store.Run, error) {
	rows, err := d.db.Query(`
		SELECT id, task_id, rc_job_id, status, trigger, summary, error, created_at, updated_at
		FROM runs ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var runs []store.Run
	for rows.Next() {
		var r store.Run
		var summaryJSON string
		if err := rows.Scan(&r.ID, &r.TaskID, &r.RcJobID, &r.Status, &r.Trigger, &summaryJSON, &r.Error, &r.CreatedAt, &r.UpdatedAt); err != nil {
			continue
		}
		if summaryJSON != "" && summaryJSON != "{}" {
			json.Unmarshal([]byte(summaryJSON), &r.Summary)
		}
		runs = append(runs, r)
	}
	return runs, nil
}

// GetByTaskID 根据任务ID获取运行记录
func (d *RunDAO) GetByTaskID(taskID int64) ([]store.Run, error) {
	rows, err := d.db.Query(`
		SELECT id, task_id, rc_job_id, status, trigger, summary, error, created_at, updated_at
		FROM runs WHERE task_id = ? ORDER BY created_at DESC`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var runs []store.Run
	for rows.Next() {
		var r store.Run
		var summaryJSON string
		if err := rows.Scan(&r.ID, &r.TaskID, &r.RcJobID, &r.Status, &r.Trigger, &summaryJSON, &r.Error, &r.CreatedAt, &r.UpdatedAt); err != nil {
			continue
		}
		if summaryJSON != "" && summaryJSON != "{}" {
			json.Unmarshal([]byte(summaryJSON), &r.Summary)
		}
		runs = append(runs, r)
	}
	return runs, nil
}

// GetRunning 获取运行中的任务
func (d *RunDAO) GetRunning() ([]store.JobStatus, error) {
	rows, err := d.db.Query(`
		SELECT id, rc_job_id, status, summary, error 
		FROM runs 
		WHERE status = 'running' AND rc_job_id > 0
		ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var runs []store.JobStatus
	for rows.Next() {
		var r store.JobStatus
		var summaryJSON string
		if err := rows.Scan(&r.ID, &r.RcJobID, &r.Status, &summaryJSON, &r.Error); err != nil {
			continue
		}
		if summaryJSON != "" && summaryJSON != "{}" {
			json.Unmarshal([]byte(summaryJSON), &r.Summary)
		}
		runs = append(runs, r)
	}
	return runs, nil
}

// Update 更新运行记录
func (d *RunDAO) Update(id int64, updateFn func(*store.Run)) error {
	r, err := d.GetByID(id)
	if err != nil {
		return nil
	}
	
	updateFn(&r)
	
	summaryBytes, _ := json.Marshal(r.Summary)
	_, err = d.db.Exec(`
		UPDATE runs SET status = ?, summary = ?, error = ?, updated_at = datetime('now')
		WHERE id = ?`,
		r.Status, string(summaryBytes), r.Error, id)
	return err
}

// UpdateStatus 更新运行状态
func (d *RunDAO) UpdateStatus(id int64, status, errorMsg string, summary map[string]any) error {
	summaryBytes, _ := json.Marshal(summary)
	finishedAt := time.Now()
	
	_, err := d.db.Exec(`
		UPDATE runs SET status = ?, summary = ?, error = ?, finished_at = ?
		WHERE id = ?`,
		status, string(summaryBytes), errorMsg, finishedAt, id)
	return err
}

// Delete 删除运行记录
func (d *RunDAO) Delete(id int64) error {
	_, err := d.db.Exec("DELETE FROM runs WHERE id = ?", id)
	return err
}
