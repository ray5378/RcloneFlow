package store

import (
	"time"
)

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

func (db *DB) DeleteSchedule(id int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.db.Exec("DELETE FROM schedules WHERE id = ?", id)
	return err
}

func (db *DB) SetScheduleEnabled(id int64, enabled bool) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.db.Exec("UPDATE schedules SET enabled = ? WHERE id = ?", enabled, id)
	return err
}

func (db *DB) UpdateScheduleNextRunTime(id int64, nextRunTime time.Time) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.db.Exec("UPDATE schedules SET next_run_time = ? WHERE id = ?", nextRunTime, id)
	return err
}

func (db *DB) UpdateScheduleSpec(id int64, spec string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	_, err := db.db.Exec("UPDATE schedules SET spec = ? WHERE id = ?", spec, id)
	return err
}