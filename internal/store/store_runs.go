package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func (db *DB) ListRuns(page, pageSize int) ([]Run, int, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var total int
	err := db.db.QueryRow(`SELECT COUNT(*) FROM runs`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	rows, err := db.db.Query(`
		SELECT id, task_id, status, trigger, summary, error, created_at, updated_at,
		       task_name, task_mode, source_remote, source_path, target_remote, target_path, finished_at, bytes_transferred, speed
		FROM runs ORDER BY id DESC LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	runs, err := db.scanRuns(rows)
	return runs, total, err
}

func (db *DB) ListRunsByTask(taskID int64) ([]Run, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	rows, err := db.db.Query(`
		SELECT id, task_id, status, trigger, summary, error, created_at, updated_at,
		       task_name, task_mode, source_remote, source_path, target_remote, target_path, finished_at, bytes_transferred, speed
		FROM runs WHERE task_id = ? ORDER BY id DESC`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return db.scanRuns(rows)
}

func (db *DB) ListActiveRuns() ([]Run, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	rows, err := db.db.Query(`
		SELECT id, task_id, status, trigger, summary, error, created_at, updated_at,
		       task_name, task_mode, source_remote, source_path, target_remote, target_path, finished_at, bytes_transferred, speed
		FROM runs WHERE status IN ('running', 'finalizing') ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return db.scanRuns(rows)
}

func (db *DB) ClearAllRunningStatus() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.db.Exec(`UPDATE runs SET status = 'stopped', finished_at = datetime('now') WHERE status IN ('running', 'finalizing')`)
	return err
}

func (db *DB) TryAcquireRun(run *Run) (*Run, bool, error) {
	tx, err := db.db.Begin()
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback()

	var count int
	err = tx.QueryRow(`SELECT COUNT(*) FROM runs WHERE status = 'running'`).Scan(&count)
	if err != nil {
		return nil, false, err
	}
	if count > 0 {
		return nil, true, nil
	}

	summaryJSON, err := json.Marshal(run.Summary)
	if err != nil {
		return nil, false, err
	}

	_, err = tx.Exec(`
		INSERT INTO runs (task_id, status, trigger, summary, error, created_at, updated_at,
		                 task_name, task_mode, source_remote, source_path, target_remote, target_path, bytes_transferred, speed)
		VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'), ?, ?, ?, ?, ?, ?, 0, '')`,
		run.TaskID, run.Status, run.Trigger, string(summaryJSON), run.Error,
		run.TaskName, run.TaskMode, run.SourceRemote, run.SourcePath, run.TargetRemote, run.TargetPath)
	if err != nil {
		return nil, false, err
	}

	var id int64
	err = tx.QueryRow(`SELECT last_insert_rowid()`).Scan(&id)
	if err != nil {
		return nil, false, err
	}
	run.ID = id

	if err = tx.Commit(); err != nil {
		return nil, false, err
	}
	return run, false, nil
}

func (db *DB) GetActiveRunByTaskID(taskID int64) (Run, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var r Run
	var summaryJSON string
	var finishedAt sql.NullTime
	var speed sql.NullString
	var taskName, taskMode, sourceRemote, sourcePath, targetRemote, targetPath sql.NullString
	var bytesTransferred sql.NullInt64
	err := db.db.QueryRow(`
		SELECT id, task_id, status, trigger, summary, error, created_at, updated_at,
		       task_name, task_mode, source_remote, source_path, target_remote, target_path, finished_at, bytes_transferred, speed
		FROM runs WHERE task_id = ? AND status IN ('running', 'finalizing')
		ORDER BY created_at DESC LIMIT 1`, taskID).Scan(
		&r.ID, &r.TaskID, &r.Status, &r.Trigger, &summaryJSON, &r.Error, &r.CreatedAt, &r.UpdatedAt,
		&taskName, &taskMode, &sourceRemote, &sourcePath, &targetRemote, &targetPath,
		&finishedAt, &bytesTransferred, &speed)
	if err != nil {
		return Run{}, fmt.Errorf("active run not found for task %d", taskID)
	}
	if err := json.Unmarshal([]byte(summaryJSON), &r.Summary); err != nil {
		r.Summary = make(map[string]any)
	}
	if finishedAt.Valid {
		r.FinishedAt = &finishedAt.Time
	}
	if taskName.Valid {
		r.TaskName = taskName.String
	}
	if taskMode.Valid {
		r.TaskMode = taskMode.String
	}
	if sourceRemote.Valid {
		r.SourceRemote = sourceRemote.String
	}
	if sourcePath.Valid {
		r.SourcePath = sourcePath.String
	}
	if targetRemote.Valid {
		r.TargetRemote = targetRemote.String
	}
	if targetPath.Valid {
		r.TargetPath = targetPath.String
	}
	if bytesTransferred.Valid {
		r.BytesTransferred = bytesTransferred.Int64
	}
	if speed.Valid {
		r.Speed = speed.String
	}
	if r.FinishedAt == nil && r.Summary != nil {
		if s, ok := r.Summary["finishedAt"].(string); ok && s != "" {
			if t, e := time.Parse(time.RFC3339, s); e == nil {
				r.FinishedAt = &t
			}
		}
	}
	return r, nil
}

func (db *DB) scanRuns(rows *sql.Rows) ([]Run, error) {
	var runs []Run
	for rows.Next() {
		var r Run
		var summaryJSON string
		var finishedAt sql.NullTime
		var speed sql.NullString
		var taskName, taskMode, sourceRemote, sourcePath, targetRemote, targetPath sql.NullString
		var bytesTransferred sql.NullInt64
		err := rows.Scan(&r.ID, &r.TaskID, &r.Status, &r.Trigger, &summaryJSON, &r.Error, &r.CreatedAt, &r.UpdatedAt,
			&taskName, &taskMode, &sourceRemote, &sourcePath, &targetRemote, &targetPath,
			&finishedAt, &bytesTransferred, &speed)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(summaryJSON), &r.Summary); err != nil {
			r.Summary = make(map[string]any)
		}
		if finishedAt.Valid {
			r.FinishedAt = &finishedAt.Time
		}
		if taskName.Valid {
			r.TaskName = taskName.String
		}
		if taskMode.Valid {
			r.TaskMode = taskMode.String
		}
		if sourceRemote.Valid {
			r.SourceRemote = sourceRemote.String
		}
		if sourcePath.Valid {
			r.SourcePath = sourcePath.String
		}
		if targetRemote.Valid {
			r.TargetRemote = targetRemote.String
		}
		if targetPath.Valid {
			r.TargetPath = targetPath.String
		}
		if bytesTransferred.Valid {
			r.BytesTransferred = bytesTransferred.Int64
		}
		if speed.Valid {
			r.Speed = speed.String
		}
		if r.FinishedAt == nil && r.Summary != nil {
			if s, ok := r.Summary["finishedAt"].(string); ok && s != "" {
				if t, e := time.Parse(time.RFC3339, s); e == nil {
					r.FinishedAt = &t
				}
			}
		}
		runs = append(runs, r)
	}
	return runs, nil
}

func (db *DB) AddRun(r Run) (Run, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	summaryJSON, err := json.Marshal(r.Summary)
	if err != nil {
		summaryJSON = []byte("{}")
	}

	result, err := db.db.Exec(`
		INSERT INTO runs (task_id, status, trigger, summary, error, task_name, task_mode, source_remote, source_path, target_remote, target_path) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.TaskID, r.Status, r.Trigger, string(summaryJSON), r.Error,
		r.TaskName, r.TaskMode, r.SourceRemote, r.SourcePath, r.TargetRemote, r.TargetPath)
	if err != nil {
		return Run{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Run{}, err
	}

	r.ID = id
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	return r, nil
}

func (db *DB) UpdateRun(id int64, fn func(*Run)) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	var r Run
	var summaryJSON string
	err := db.db.QueryRow(`
		SELECT id, task_id, status, trigger, summary, error, created_at, updated_at 
		FROM runs WHERE id = ?`, id).Scan(
		&r.ID, &r.TaskID, &r.Status, &r.Trigger, &summaryJSON, &r.Error, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(summaryJSON), &r.Summary); err != nil {
		r.Summary = make(map[string]any)
	}

	fn(&r)
	r.UpdatedAt = time.Now()

	summaryBytes, err := json.Marshal(r.Summary)
	if err != nil {
		return fmt.Errorf("marshal run summary: %w", err)
	}

	_, err = db.db.Exec(`
		UPDATE runs SET status = ?, summary = ?, error = ?, updated_at = ? WHERE id = ?`,
		r.Status, string(summaryBytes), r.Error, r.UpdatedAt, id)
	return err
}

func (db *DB) GetRun(id int64) (Run, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var r Run
	var summaryJSON string
	err := db.db.QueryRow(`
		SELECT id, task_id, status, trigger, summary, error, created_at, updated_at 
		FROM runs WHERE id = ?`, id).Scan(
		&r.ID, &r.TaskID, &r.Status, &r.Trigger, &summaryJSON, &r.Error, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return Run{}, fmt.Errorf("get run %d: %w", id, err)
	}
	if err := json.Unmarshal([]byte(summaryJSON), &r.Summary); err != nil {
		r.Summary = make(map[string]any)
	}
	return r, nil
}

func (db *DB) UpdateRunStatus(id int64, status, errorMsg string, summary map[string]any) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	summaryBytes, err := json.Marshal(summary)
	if err != nil {
		return fmt.Errorf("marshal run summary: %w", err)
	}
	finishedAt := time.Now()

	var execErr error
	_, execErr = db.db.Exec(`
		UPDATE runs SET status = ?, summary = ?, error = ?, updated_at = ?, finished_at = ?
		WHERE id = ?`,
		status, string(summaryBytes), errorMsg, finishedAt, finishedAt, id)
	return execErr
}

func (db *DB) UpdateRunProgress(id int64, data float64, detail string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	newBytes := int64(data)
	_, err := db.db.Exec(`
		UPDATE runs SET bytes_transferred = CASE WHEN ? > bytes_transferred THEN ? ELSE bytes_transferred END, speed = ?, updated_at = ?
		WHERE id = ?`,
		newBytes, newBytes, detail, time.Now(), id)
	return err
}

func (db *DB) Vacuum() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	_, err := db.db.Exec("VACUUM")
	return err
}

func (db *DB) DeleteRunsByIDs(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	db.mu.Lock()
	defer db.mu.Unlock()

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	query := fmt.Sprintf("DELETE FROM runs WHERE id IN (%s)", strings.Join(placeholders, ","))
	_, err := db.db.Exec(query, args...)
	return err
}

func (db *DB) DeleteRun(id int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.db.Exec("DELETE FROM runs WHERE id = ?", id)
	return err
}

func (db *DB) DeleteAllRuns() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.db.Exec("DELETE FROM runs")
	return err
}

func (db *DB) DeleteRunsByTask(taskId int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.db.Exec("DELETE FROM runs WHERE task_id = ?", taskId)
	return err
}

func (db *DB) CleanOldRuns(days int) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	result, err := db.db.Exec("DELETE FROM runs WHERE created_at < datetime('now', ?)", fmt.Sprintf("-%d days", days))
	if err != nil {
		return 0, err
	}
	n, _ := result.RowsAffected()
	return int(n), nil
}