package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v4"
	_ "github.com/pressly/goose/v4/drivers/sqlite3"
)

var db *sql.DB

// InitDB 初始化数据库连接
func InitDB(dir, dsn string) (*sql.DB, error) {
	// 确保目录存在
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据库目录失败: %w", err)
	}

	// 连接数据库
	var err error
	db, err = sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	// 设置连接池
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// 启用 WAL 模式
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("设置WAL模式失败: %w", err)
	}

	// 启用外键约束
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("启用外键约束失败: %w", err)
	}

	return db, nil
}

// RunMigrations 运行数据库迁移
func RunMigrations(db *sql.DB, migrationsDir string) error {
	// 确保迁移目录存在
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		// 如果迁移目录不存在，使用内联迁移
		return nil
	}

	// 设置数据库
	goose.SetDB(db)

	// 设置 goose
	goose.SetVerbose(true)
	goose.SetGlobal(true)

	// 运行迁移
	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("运行迁移失败: %w", err)
	}

	return nil
}

// GetDB 获取全局数据库实例
func GetDB() *sql.DB {
	return db
}

// GetMigrationsDir 获取迁移目录路径
func GetMigrationsDir() string {
	// 优先使用环境变量
	if dir := os.Getenv("MIGRATIONS_DIR"); dir != "" {
		return dir
	}
	
	// 尝试从当前目录向上查找
	dir, _ := filepath.Split(os.Args[0])
	if dir != "" {
		migrationsPath := filepath.Join(dir, "migrations")
		if _, err := os.Stat(migrationsPath); err == nil {
			return migrationsPath
		}
	}
	
	// 默认当前目录
	return "./migrations"
}
