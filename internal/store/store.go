package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// User 用户模型
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // 不返回密码
	CreatedAt time.Time `json:"createdAt"`
}

type Task struct {
	ID           int64           `json:"id"`
	Name         string          `json:"name"`
	Mode         string          `json:"mode"`
	SourceRemote string          `json:"sourceRemote"`
	SourcePath   string          `json:"sourcePath"`
	TargetRemote string          `json:"targetRemote"`
	TargetPath   string          `json:"targetPath"`
	SortOrder    int64           `json:"sortOrder"`
	Options      json.RawMessage `json:"options,omitempty"`
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

type DB struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewDB 创建数据库实例
func NewDB(db *sql.DB) *DB {
	return &DB{db: db}
}

func Open(dir string) (*DB, error) {
	_ = os.MkdirAll(dir, 0o755)
	path := filepath.Join(dir, "rcloneflow.db")

	db, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// SQLite 单文件库在本项目里优先追求稳定性而不是并发吞吐；
	// 主运行时链路收紧为单连接，减少同进程内并发写导致的 busy/locked 风险。
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	s := NewDB(db)
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return s, nil
}

// openDB 打开数据库（供内部迁移使用）
func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	// 内部迁移辅助连接也收紧成单连接，避免与主链形成不同的并发策略。
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) migrate() error {
	// 创建版本表（如果不存在）
	_, _ = db.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)

	// 定义迁移
	type migration struct {
		version int
		sql     string
	}

	migrations := []migration{
		{
			version: 1,
			sql: `
				CREATE TABLE IF NOT EXISTS tasks (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					name TEXT NOT NULL,
					mode TEXT NOT NULL,
					source_remote TEXT NOT NULL,
					source_path TEXT NOT NULL,
					target_remote TEXT NOT NULL,
					target_path TEXT NOT NULL,
					options TEXT,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP
				);
				
				CREATE TABLE IF NOT EXISTS schedules (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					task_id INTEGER NOT NULL,
					spec TEXT NOT NULL,
					enabled BOOLEAN DEFAULT 1,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					next_run_time DATETIME,
					FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
				);
				
				CREATE TABLE IF NOT EXISTS runs (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					task_id INTEGER NOT NULL,
					rc_job_id INTEGER DEFAULT 0,
					status TEXT NOT NULL,
					trigger TEXT NOT NULL,
					summary TEXT DEFAULT '{}',
					error TEXT DEFAULT '',
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					finished_at DATETIME,
					task_name TEXT,
					task_mode TEXT,
					source_remote TEXT,
					source_path TEXT,
					target_remote TEXT,
					target_path TEXT,
					bytes_transferred INTEGER DEFAULT 0,
					speed TEXT,
					FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
				);

				CREATE TABLE IF NOT EXISTS users (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					username TEXT UNIQUE NOT NULL,
					password TEXT NOT NULL,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP
				);
				
				CREATE INDEX IF NOT EXISTS idx_runs_task_id ON runs(task_id);
				CREATE INDEX IF NOT EXISTS idx_runs_created_at ON runs(created_at);
				CREATE INDEX IF NOT EXISTS idx_runs_status ON runs(status);
				CREATE INDEX IF NOT EXISTS idx_schedules_task_id ON schedules(task_id);
			`,
		},
		{
			version: 2,
			sql: `
				CREATE TABLE IF NOT EXISTS runs_new (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					task_id INTEGER NOT NULL,
					status TEXT NOT NULL,
					trigger TEXT NOT NULL,
					summary TEXT DEFAULT '{}',
					error TEXT DEFAULT '',
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					finished_at DATETIME,
					task_name TEXT,
					task_mode TEXT,
					source_remote TEXT,
					source_path TEXT,
					target_remote TEXT,
					target_path TEXT,
					bytes_transferred INTEGER DEFAULT 0,
					speed TEXT,
					FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
				);

				INSERT INTO runs_new (
					id, task_id, status, trigger, summary, error, created_at, updated_at,
					finished_at, task_name, task_mode, source_remote, source_path,
					target_remote, target_path, bytes_transferred, speed
				)
				SELECT
					id, task_id, status, trigger, summary, error, created_at, updated_at,
					finished_at, task_name, task_mode, source_remote, source_path,
					target_remote, target_path, bytes_transferred, speed
				FROM runs;

				DROP TABLE runs;
				ALTER TABLE runs_new RENAME TO runs;

				CREATE INDEX IF NOT EXISTS idx_runs_task_id ON runs(task_id);
				CREATE INDEX IF NOT EXISTS idx_runs_created_at ON runs(created_at);
				CREATE INDEX IF NOT EXISTS idx_runs_status ON runs(status);
			`,
		},
		{
			version: 3,
			sql: `
				CREATE UNIQUE INDEX IF NOT EXISTS idx_tasks_name_unique ON tasks(name COLLATE NOCASE);
			`,
		},
		{
			version: 4,
			sql: `
				ALTER TABLE tasks ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0;
				UPDATE tasks SET sort_order = id WHERE sort_order = 0;
				CREATE INDEX IF NOT EXISTS idx_tasks_sort_order ON tasks(sort_order, id);
			`,
		},
	}

	// 获取当前版本
	var currentVersion int
	row := db.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations")
	if err := row.Scan(&currentVersion); err != nil {
		currentVersion = 0
	}

	// 应用待处理的迁移
	for _, m := range migrations {
		if m.version <= currentVersion {
			continue
		}
		if _, err := db.db.Exec(m.sql); err != nil {
			return fmt.Errorf("应用迁移 v%d 失败: %w", m.version, err)
		}
		if _, err := db.db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", m.version); err != nil {
			return fmt.Errorf("记录迁移版本 %d 失败: %w", m.version, err)
		}
	}

	return nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

// ===== Tasks =====

func (db *DB) ListTasks() ([]Task, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	rows, err := db.db.Query(`
		SELECT id, name, mode, source_remote, source_path, target_remote, target_path, sort_order, options, created_at 
		FROM tasks ORDER BY sort_order ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		var options sql.NullString
		err := rows.Scan(&t.ID, &t.Name, &t.Mode, &t.SourceRemote, &t.SourcePath, &t.TargetRemote, &t.TargetPath, &t.SortOrder, &options, &t.CreatedAt)
		if err != nil {
			return nil, err
		}
		if options.Valid {
			t.Options = []byte(options.String)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (db *DB) AddTask(t Task) (Task, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if t.SortOrder == 0 {
		_ = db.db.QueryRow(`SELECT COALESCE(MAX(sort_order), 0) + 1 FROM tasks`).Scan(&t.SortOrder)
	}

	result, err := db.db.Exec(`
		INSERT INTO tasks (name, mode, source_remote, source_path, target_remote, target_path, sort_order, options) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		t.Name, t.Mode, t.SourceRemote, t.SourcePath, t.TargetRemote, t.TargetPath, t.SortOrder, t.Options)
	if err != nil {
		return Task{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Task{}, err
	}

	t.ID = id
	t.CreatedAt = time.Now()
	return t, nil
}

func (db *DB) GetTask(id int64) (Task, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var t Task
	var options sql.NullString
	err := db.db.QueryRow(`
		SELECT id, name, mode, source_remote, source_path, target_remote, target_path, sort_order, options, created_at 
		FROM tasks WHERE id = ?`, id).Scan(
		&t.ID, &t.Name, &t.Mode, &t.SourceRemote, &t.SourcePath, &t.TargetRemote, &t.TargetPath, &t.SortOrder, &options, &t.CreatedAt)
	if err != nil {
		return Task{}, false
	}
	if options.Valid {
		t.Options = []byte(options.String)
	}
	return t, true
}

func (db *DB) GetSchedule(id int64) (Schedule, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var s Schedule
	err := db.db.QueryRow(`
		SELECT id, task_id, spec, enabled, created_at 
		FROM schedules WHERE id = ?`, id).Scan(
		&s.ID, &s.TaskID, &s.Spec, &s.Enabled, &s.CreatedAt)
	if err != nil {
		return Schedule{}, false
	}
	return s, true
}

func (db *DB) UpdateTask(id int64, t Task) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.db.Exec(`
		UPDATE tasks SET name=?, mode=?, source_remote=?, source_path=?, target_remote=?, target_path=?, sort_order=?, options=?
		WHERE id=?`, t.Name, t.Mode, t.SourceRemote, t.SourcePath, t.TargetRemote, t.TargetPath, t.SortOrder, t.Options, id)
	return err
}

func (db *DB) UpdateTaskSortOrders(updates map[int64]int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare(`UPDATE tasks SET sort_order = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for id, sortOrder := range updates {
		if _, err = stmt.Exec(sortOrder, id); err != nil {
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

// CleanOldRuns 删除指定天数之前的运行记录
func (db *DB) CleanOldRuns(days int) (int64, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	result, err := db.db.Exec("DELETE FROM runs WHERE created_at < datetime('now', ?)", fmt.Sprintf("-%d days", days))
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}

// ===== Schedules =====

func (db *DB) ListSchedules() ([]Schedule, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	rows, err := db.db.Query(`
		SELECT id, task_id, spec, enabled, created_at 
		FROM schedules ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []Schedule
	for rows.Next() {
		var s Schedule
		err := rows.Scan(&s.ID, &s.TaskID, &s.Spec, &s.Enabled, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}
	return schedules, nil
}

func (db *DB) AddSchedule(s Schedule) (Schedule, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	result, err := db.db.Exec(`
		INSERT INTO schedules (task_id, spec, enabled) VALUES (?, ?, ?)`,
		s.TaskID, s.Spec, s.Enabled)
	if err != nil {
		return Schedule{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Schedule{}, err
	}

	s.ID = id
	s.CreatedAt = time.Now()
	return s, nil
}

func (db *DB) DeleteSchedule(id int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.db.Exec("DELETE FROM schedules WHERE id = ?", id)
	return err
}

// SetScheduleEnabled 启用/禁用定时任务
func (db *DB) SetScheduleEnabled(id int64, enabled bool) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.db.Exec("UPDATE schedules SET enabled = ? WHERE id = ?", enabled, id)
	return err
}

// UpdateScheduleNextRunTime 更新任务的下次触发时间
func (db *DB) UpdateScheduleNextRunTime(id int64, nextRunTime time.Time) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.db.Exec("UPDATE schedules SET next_run_time = ? WHERE id = ?", nextRunTime, id)
	return err
}

// UpdateScheduleSpec 更新定时表达式（不重建记录）
func (db *DB) UpdateScheduleSpec(id int64, spec string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	_, err := db.db.Exec("UPDATE schedules SET spec = ? WHERE id = ?", spec, id)
	return err
}

// ===== Runs =====

func (db *DB) ListRuns(page, pageSize int) ([]Run, int, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// 获取总数
	var total int
	err := db.db.QueryRow(`SELECT COUNT(*) FROM runs`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
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
	db.mu.Lock()
	defer db.mu.Unlock()

	rows, err := db.db.Query(`
		SELECT id, task_id, status, trigger, summary, error, created_at, updated_at,
		       task_name, task_mode, source_remote, source_path, target_remote, target_path, finished_at, bytes_transferred, speed
		FROM runs WHERE status = 'running' ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return db.scanRuns(rows)
}

// ClearAllRunningStatus 清空所有运行状态（容器重启后恢复时调用）
func (db *DB) ClearAllRunningStatus() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.db.Exec(`UPDATE runs SET status = 'stopped', finished_at = datetime('now') WHERE status = 'running'`)
	return err
}

// TryAcquireRun 尝试原子性地创建运行记录（单例模式用）
// 返回 (run, existed, error)
// existed=true 表示已有任务在运行，run=nil
// existed=false 表示成功创建，run 为新记录
func (db *DB) TryAcquireRun(run *Run) (*Run, bool, error) {
	// 使用 BEGIN EXCLUSIVE 事务确保跨实例原子性
	tx, err := db.db.Begin()
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback()

	// 检查是否有正在运行的任务
	var count int
	err = tx.QueryRow(`SELECT COUNT(*) FROM runs WHERE status = 'running'`).Scan(&count)
	if err != nil {
		return nil, false, err
	}
	if count > 0 {
		return nil, true, nil // 已有任务在运行
	}

	// 序列化 Summary 为 JSON
	summaryJSON, err := json.Marshal(run.Summary)
	if err != nil {
		return nil, false, err
	}

	// 插入新记录
	_, err = tx.Exec(`
		INSERT INTO runs (task_id, status, trigger, summary, error, created_at, updated_at,
		                 task_name, task_mode, source_remote, source_path, target_remote, target_path, bytes_transferred, speed)
		VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'), ?, ?, ?, ?, ?, ?, 0, '')`,
		run.TaskID, run.Status, run.Trigger, string(summaryJSON), run.Error,
		run.TaskName, run.TaskMode, run.SourceRemote, run.SourcePath, run.TargetRemote, run.TargetPath)
	if err != nil {
		return nil, false, err
	}

	// 获取刚插入的记录ID
	var id int64
	err = tx.QueryRow(`SELECT last_insert_rowid()`).Scan(&id)
	if err != nil {
		return nil, false, err
	}
	run.ID = id

	// 提交事务
	if err = tx.Commit(); err != nil {
		return nil, false, err
	}
	return run, false, nil
}

// GetActiveRunByTaskID 获取任务当前运行中的记录
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
		FROM runs WHERE task_id = ? AND status = 'running'
		ORDER BY created_at DESC LIMIT 1`, taskID).Scan(
		&r.ID, &r.TaskID, &r.Status, &r.Trigger, &summaryJSON, &r.Error, &r.CreatedAt, &r.UpdatedAt,
		&taskName, &taskMode, &sourceRemote, &sourcePath, &targetRemote, &targetPath,
		&finishedAt, &bytesTransferred, &speed)
	if err != nil {
		return Run{}, err
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
	// Fallback: if DB finished_at empty, try summary.finishedAt
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
		// Fallback: if DB finished_at is empty, use summary.finishedAt when available
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

	// Fetch current
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

	// Apply update
	fn(&r)
	r.UpdatedAt = time.Now()

	summaryBytes, _ := json.Marshal(r.Summary)

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
		return Run{}, err
	}
	if err := json.Unmarshal([]byte(summaryJSON), &r.Summary); err != nil {
		r.Summary = make(map[string]any)
	}
	return r, nil
}

// UpdateRunStatus 更新运行状态
func (db *DB) UpdateRunStatus(id int64, status, errorMsg string, summary map[string]any) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	summaryBytes, _ := json.Marshal(summary)
	finishedAt := time.Now()

	_, err := db.db.Exec(`
		UPDATE runs SET status = ?, summary = ?, error = ?, updated_at = ?, finished_at = ?
		WHERE id = ?`,
		status, string(summaryBytes), errorMsg, finishedAt, finishedAt, id)
	return err
}

// UpdateRunProgress 更新运行进度（bytes和speed），只有新进度大于旧进度时才更新，防止回退
func (db *DB) UpdateRunProgress(id int64, bytesTransferred int64, speed string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// 先查询当前进度，只有新进度大于旧进度时才更新
	var currentBytes int64
	err := db.db.QueryRow("SELECT bytes_transferred FROM runs WHERE id = ?", id).Scan(&currentBytes)
	if err != nil {
		return err
	}
	if bytesTransferred > currentBytes {
		_, err = db.db.Exec(`
			UPDATE runs SET bytes_transferred = ?, speed = ?, updated_at = ?
			WHERE id = ?`,
			bytesTransferred, speed, time.Now(), id)
	}
	return err
}

// ===== 用户相关操作 =====

// CreateUser 创建用户
func (db *DB) CreateUser(username, password string) (User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	result, err := db.db.Exec(`
		INSERT INTO users (username, password) VALUES (?, ?)`,
		username, password)
	if err != nil {
		return User{}, err
	}

	id, _ := result.LastInsertId()
	return User{
		ID:        id,
		Username:  username,
		Password:  password,
		CreatedAt: time.Now(),
	}, nil
}

// GetUserByUsername 根据用户名获取用户
func (db *DB) GetUserByUsername(username string) (User, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var u User
	err := db.db.QueryRow(`
		SELECT id, username, password, created_at FROM users WHERE username = ?`, username).
		Scan(&u.ID, &u.Username, &u.Password, &u.CreatedAt)
	if err != nil {
		return User{}, false
	}
	return u, true
}

// GetUserByID 根据ID获取用户
func (db *DB) GetUserByID(id int64) (User, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var u User
	err := db.db.QueryRow(`
		SELECT id, username, password, created_at FROM users WHERE id = ?`, id).
		Scan(&u.ID, &u.Username, &u.Password, &u.CreatedAt)
	if err != nil {
		return User{}, false
	}
	return u, true
}

// UpdatePassword 更新密码
func (db *DB) UpdatePassword(id int64, hashedPassword string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.db.Exec(`UPDATE users SET password = ? WHERE id = ?`, hashedPassword, id)
	return err
}

// UpdateUsername 更新用户名
func (db *DB) UpdateUsername(id int64, username string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.db.Exec(`UPDATE users SET username = ? WHERE id = ?`, username, id)
	return err
}

// ListUsers 获取所有用户
func (db *DB) ListUsers() ([]User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	rows, err := db.db.Query(`SELECT id, username, password, created_at FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Password, &u.CreatedAt); err != nil {
			continue
		}
		users = append(users, u)
	}
	return users, nil
}
