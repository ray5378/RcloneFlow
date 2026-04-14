-- +goose Up
-- +goose StatementBegin
-- 添加 finished_at 字段（如果不存在）
-- SQLite 不支持直接 ADD COLUMN IF NOT EXISTS，所以需要用这种方式
-- 如果字段已存在会报错，但迁移不会中断
ALTER TABLE runs ADD COLUMN finished_at DATETIME;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- SQLite 不支持 DROP COLUMN，所以这个回滚是空的
-- 在生产环境中，如果需要回滚，需要重建表
-- +goose StatementEnd
