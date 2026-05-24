-- 添加 bisync_options 列存储 bisync 特定配置
ALTER TABLE tasks ADD COLUMN bisync_options TEXT;
