package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *FsController) HandleMkdir(w http.ResponseWriter, r *http.Request) { c.wrap(w, r, c.doMkdir) }

func (c *FsController) doMkdir(ctx context.Context, body []byte) (any, error) {
	var req fileOpReq
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("无效的请求格式: %w", err)
	}
	fs, p := normalize(sanitizeFsRemote(req.Fs), sanitizePath(req.Remote))
	_, err := runRclone(ctx, "mkdir", fs+p)
	return nil, err
}