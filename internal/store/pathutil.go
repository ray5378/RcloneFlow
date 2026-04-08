package store

import (
	"os"
	"path/filepath"
)

// PathJoin 返回数据目录下的子路径（用于本地运行时配置文件等）
func (db *DB) PathJoin(rel string) string {
	// 兼容：默认数据目录位于进程 cwd 下的 data/
	cwd, _ := os.Getwd()
	p := filepath.Join(cwd, "data", rel)
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	return p
}
