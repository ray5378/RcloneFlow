package store

import (
	"database/sql"
	"time"
)

func (db *DB) ListTasks() ([]Task, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	rows, err := db.db.Query(`
		SELECT id, name, mode, source_remote, source_path, target_remote, target_path, options, bisync_options, sort_order, created_at 
		FROM tasks ORDER BY sort_order ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		var options sql.NullString
		var bisyncOptions sql.NullString
		err := rows.Scan(&t.ID, &t.Name, &t.Mode, &t.SourceRemote, &t.SourcePath, &t.TargetRemote, &t.TargetPath, &options, &bisyncOptions, &t.SortOrder, &t.CreatedAt)
		if err != nil {
			return nil, err
		}
		if options.Valid {
			t.Options = []byte(options.String)
		}
		if bisyncOptions.Valid {
			t.BisyncOptions = []byte(bisyncOptions.String)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (db *DB) AddTask(t Task) (Task, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var nextSortOrder int64
	if err := db.db.QueryRow(`SELECT COALESCE(MAX(sort_order), 0) + 1 FROM tasks`).Scan(&nextSortOrder); err != nil {
		return Task{}, err
	}

	result, err := db.db.Exec(`
		INSERT INTO tasks (name, mode, source_remote, source_path, target_remote, target_path, options, bisync_options, sort_order) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.Name, t.Mode, t.SourceRemote, t.SourcePath, t.TargetRemote, t.TargetPath, t.Options, t.BisyncOptions, nextSortOrder)
	if err != nil {
		return Task{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Task{}, err
	}

	t.ID = id
	t.SortOrder = nextSortOrder
	t.CreatedAt = time.Now()
	return t, nil
}

func (db *DB) GetTask(id int64) (Task, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var t Task
	var options sql.NullString
	var bisyncOptions sql.NullString
	err := db.db.QueryRow(`
		SELECT id, name, mode, source_remote, source_path, target_remote, target_path, options, bisync_options, sort_order, created_at 
		FROM tasks WHERE id = ?`, id).Scan(
		&t.ID, &t.Name, &t.Mode, &t.SourceRemote, &t.SourcePath, &t.TargetRemote, &t.TargetPath, &options, &bisyncOptions, &t.SortOrder, &t.CreatedAt)
	if err != nil {
		return Task{}, false
	}
	if options.Valid {
		t.Options = []byte(options.String)
	}
	if bisyncOptions.Valid {
		t.BisyncOptions = []byte(bisyncOptions.String)
	}
	return t, true
}

func (db *DB) UpdateTask(id int64, t Task) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.db.Exec(`
		UPDATE tasks SET name=?, mode=?, source_remote=?, source_path=?, target_remote=?, target_path=?, options=?, bisync_options=?, sort_order=?
		WHERE id=?`, t.Name, t.Mode, t.SourceRemote, t.SourcePath, t.TargetRemote, t.TargetPath, t.Options, t.BisyncOptions, t.SortOrder, id)
	return err
}

func (db *DB) UpdateTaskSortOrders(updates map[int64]int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`UPDATE tasks SET sort_order = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for taskID, sortOrder := range updates {
		if _, err := stmt.Exec(sortOrder, taskID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) DeleteTask(id int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}