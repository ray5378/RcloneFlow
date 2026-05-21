package controller

import (
	"net/http"
	"path"
	"strings"

	"rcloneflow/internal/adapter"
)

// BrowserController 文件浏览器控制器
type BrowserController struct {
	rc *adapter.RcloneClient
}

// NewBrowserController 创建文件浏览器控制器
func NewBrowserController(rc *adapter.RcloneClient) *BrowserController {
	return &BrowserController{rc: rc}
}

// HandleList 处理文件列表请求
func (c *BrowserController) HandleList(w http.ResponseWriter, r *http.Request) {
	remote := r.URL.Query().Get("remote")
	p := strings.Trim(strings.TrimPrefix(r.URL.Query().Get("path"), "/"), " ")
	if p == "." {
		p = ""
	}

	fsPath := remote + ":"
	items, err := c.rc.ListPath(r.Context(), fsPath, p)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}

	result := make([]map[string]any, len(items))
	for i, item := range items {
		result[i] = map[string]any{
			"Name":     item.Name,
			"Path":     item.Path,
			"IsDir":    item.IsDir,
			"MimeType": item.MimeType,
			"ModTime":  item.ModTime,
			"Size":     item.Size,
		}
	}

	current := fsPath
	if p != "" {
		current += path.Clean("/" + p)
	}

	WriteJSON(w, 200, map[string]any{"fs": current, "items": result})
}
