package store

import (
	"database/sql"
)

// Raw 返回底层 *sql.DB（用于极少数需要原生 SQL 的地方）。
func (db *DB) Raw() *sql.DB { return db.db }
