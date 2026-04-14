package store

import (
	"database/sql"
	"fmt"
	"os"
)

var globalDB *sql.DB

// InitDB 初始化数据库连接
func InitDB(dir, dsn string) (*sql.DB, error) {
	// 确保目录存在
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据库目录失败: %w", err)
	}

	// 连接数据库
	var err error
	globalDB, err = sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 测试连接
	if err := globalDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	// 设置连接池
	globalDB.SetMaxOpenConns(25)
	globalDB.SetMaxIdleConns(5)

	// 启用 WAL 模式
	if _, err := globalDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("设置WAL模式失败: %w", err)
	}

	// 启用外键约束
	if _, err := globalDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("启用外键约束失败: %w", err)
	}

	return globalDB, nil
}

// RunMigrations 运行数据库迁移
// 当前使用内联迁移（见 store.go 中的 migrate() 方法）
func RunMigrations(db *sql.DB, migrationsDir string) error {
	// TODO: 集成 goose 迁移工具
	// 当前使用 store.go 中的内联迁移
	return nil
}

// GetDB 获取全局数据库实例
func GetDB() *sql.DB {
	return globalDB
}
