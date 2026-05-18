-- 添加 task_options 列存储 rclone 高级参数
ALTER TABLE tasks ADD COLUMN options TEXT;
