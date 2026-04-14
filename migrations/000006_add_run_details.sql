-- +goose Up
-- +goose StatementBegin
ALTER TABLE runs ADD COLUMN task_name TEXT;
ALTER TABLE runs ADD COLUMN task_mode TEXT;
ALTER TABLE runs ADD COLUMN source_remote TEXT;
ALTER TABLE runs ADD COLUMN source_path TEXT;
ALTER TABLE runs ADD COLUMN target_remote TEXT;
ALTER TABLE runs ADD COLUMN target_path TEXT;
ALTER TABLE runs ADD COLUMN finished_at DATETIME;
ALTER TABLE runs ADD COLUMN bytes_transferred INTEGER DEFAULT 0;
ALTER TABLE runs ADD COLUMN speed TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE runs DROP COLUMN task_name;
ALTER TABLE runs DROP COLUMN task_mode;
ALTER TABLE runs DROP COLUMN source_remote;
ALTER TABLE runs DROP COLUMN source_path;
ALTER TABLE runs DROP COLUMN target_remote;
ALTER TABLE runs DROP COLUMN target_path;
ALTER TABLE runs DROP COLUMN finished_at;
ALTER TABLE runs DROP COLUMN bytes_transferred;
ALTER TABLE runs DROP COLUMN speed;
-- +goose StatementEnd
