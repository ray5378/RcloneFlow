package controller

import (
	"net/http"
)

// FsController 文件系统操作控制器
// 注意：FS 传输类操作均已切换为 CLI，在 fs_cli.go 实现。
type FsController struct {}

// NewFsController 创建文件系统控制器
func NewFsController(_ any) *FsController { return &FsController{} }
