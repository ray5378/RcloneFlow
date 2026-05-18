-- +goose Up
-- +goose StatementBegin
ALTER TABLE schedules ADD COLUMN next_run_time DATETIME;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE schedules DROP COLUMN next_run_time;
-- +goose StatementEnd
