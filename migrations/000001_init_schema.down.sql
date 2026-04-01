-- 回滚初始化数据库表结构

DROP INDEX IF EXISTS idx_runs_status;
DROP INDEX IF EXISTS idx_runs_task_id;
DROP INDEX IF EXISTS idx_schedules_task_id;

DROP TABLE IF EXISTS runs;
DROP TABLE IF EXISTS schedules;
DROP TABLE IF EXISTS tasks;
