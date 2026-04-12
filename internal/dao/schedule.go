package dao

import (
	"database/sql"
	"time"

	"rcloneflow/internal/store"
)

// ScheduleDAO 定时任务数据访问对象
type ScheduleDAO struct {
	db *sql.DB
}

// NewScheduleDAO 创建ScheduleDAO
func NewScheduleDAO(db *sql.DB) *ScheduleDAO {
	return &ScheduleDAO{db: db}
}

// Create 创建定时任务
func (d *ScheduleDAO) Create(schedule store.Schedule) (store.Schedule, error) {
	result, err := d.db.Exec(`
		INSERT INTO schedules (task_id, spec, enabled, next_run_time)
		VALUES (?, ?, ?, ?)`,
		schedule.TaskID, schedule.Spec, schedule.Enabled, schedule.NextRunTime)
	if err != nil {
		return store.Schedule{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return store.Schedule{}, err
	}

	return d.GetByID(id)
}

// GetByID 根据ID获取定时任务
func (d *ScheduleDAO) GetByID(id int64) (store.Schedule, error) {
	var s store.Schedule
	var nextRunTime sql.NullTime
	err := d.db.QueryRow(`
		SELECT id, task_id, spec, enabled, next_run_time, created_at 
		FROM schedules WHERE id = ?`, id).Scan(
		&s.ID, &s.TaskID, &s.Spec, &s.Enabled, &nextRunTime, &s.CreatedAt)
	if err != nil {
		return store.Schedule{}, err
	}
	if nextRunTime.Valid {
		s.NextRunTime = &nextRunTime.Time
	}
	return s, nil
}

// GetAll 获取所有定时任务
func (d *ScheduleDAO) GetAll() ([]store.Schedule, error) {
	rows, err := d.db.Query(`
		SELECT id, task_id, spec, enabled, next_run_time, created_at 
		FROM schedules ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []store.Schedule
	for rows.Next() {
		var s store.Schedule
		var nextRunTime sql.NullTime
		if err := rows.Scan(&s.ID, &s.TaskID, &s.Spec, &s.Enabled, &nextRunTime, &s.CreatedAt); err != nil {
			continue
		}
		if nextRunTime.Valid {
			s.NextRunTime = &nextRunTime.Time
		}
		schedules = append(schedules, s)
	}
	return schedules, nil
}

// GetByTaskID 根据任务ID获取定时任务
func (d *ScheduleDAO) GetByTaskID(taskID int64) ([]store.Schedule, error) {
	rows, err := d.db.Query(`
		SELECT id, task_id, spec, enabled, next_run_time, created_at 
		FROM schedules WHERE task_id = ?`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []store.Schedule
	for rows.Next() {
		var s store.Schedule
		var nextRunTime sql.NullTime
		if err := rows.Scan(&s.ID, &s.TaskID, &s.Spec, &s.Enabled, &nextRunTime, &s.CreatedAt); err != nil {
			continue
		}
		if nextRunTime.Valid {
			s.NextRunTime = &nextRunTime.Time
		}
		schedules = append(schedules, s)
	}
	return schedules, nil
}

// UpdateNextRunTime 更新下次触发时间
func (d *ScheduleDAO) UpdateNextRunTime(id int64, nextRunTime time.Time) error {
	_, err := d.db.Exec(`
		UPDATE schedules SET next_run_time = ? WHERE id = ?`,
		nextRunTime, id)
	return err
}

// Delete 删除定时任务
func (d *ScheduleDAO) Delete(id int64) error {
	_, err := d.db.Exec("DELETE FROM schedules WHERE id = ?", id)
	return err
}
