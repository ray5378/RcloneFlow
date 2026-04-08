package controller

import (
	"net/http"
	"path"
	"strings"

	"rcloneflow/internal/rclone"
)

// BrowserController 文件浏览器控制器
type BrowserController struct {
	rc *rclone.Client
}

// NewBrowserController 创建文件浏览器控制器
func NewBrowserController(rc *rclone.Client) *BrowserController {
	return &BrowserController{rc: rc}
}

// HandleList 处理文件列表请求（CLI 等价实现：lsjson）
func (c *BrowserController) HandleList(w http.ResponseWriter, r *http.Request) {
	remote := r.URL.Query().Get("remote")
	p := strings.Trim(strings.TrimPrefix(r.URL.Query().Get("path"), "/"), " ")
	if p == "." {
		p = ""
	}

	fs := remote // 形如 "gdrive"
	entries, err := rclone.LsJSON(fs, p)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	items := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		items = append(items, map[string]any{
			"Name":  e.Name,
			"Path":  e.Path,
			"IsDir": e.IsDir,
			"Size":  e.Size,
		})
	}

	current := remote + ":"
	if p != "" {
		current += path.Clean("/" + p)
	}
	WriteJSON(w, 200, map[string]any{"fs": current, "items": items})
}
