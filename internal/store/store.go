package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

type DB struct {
	db *sql.DB
	mu sync.RWMutex
}

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

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

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

func (db *DB) migrate() error {
	_, _ = db.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)

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

				WITH ordered AS (
					SELECT id, ROW_NUMBER() OVER (ORDER BY id DESC) AS rn
					FROM tasks
				)
				UPDATE tasks
				SET sort_order = (
					SELECT rn FROM ordered WHERE ordered.id = tasks.id
				)
				WHERE sort_order = 0;
			`,
		},
	}

	var currentVersion int
	row := db.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations")
	if err := row.Scan(&currentVersion); err != nil {
		currentVersion = 0
	}

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