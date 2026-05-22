package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (c *FsController) HandleDeleteFile(w http.ResponseWriter, r *http.Request) { c.wrap(w, r, c.doDeleteFile) }
func (c *FsController) HandlePurge(w http.ResponseWriter, r *http.Request)      { c.wrap(w, r, c.doPurge) }

func (c *FsController) doDeleteFile(ctx context.Context, body []byte) (any, error) {
	var req fileOpReq
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("无效的请求格式: %w", err)
	}
	fs, p := normalize(sanitizeFsRemote(req.Fs), sanitizePath(req.Remote))
	_, err := runRclone(ctx, "deletefile", fs+p)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") { return nil, nil }
	}
	return nil, err
}

func (c *FsController) doPurge(ctx context.Context, body []byte) (any, error) {
	var req fileOpReq
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("无效的请求格式: %w", err)
	}
	fs, p := normalize(sanitizeFsRemote(req.Fs), sanitizePath(req.Remote))
	_, err := runRclone(ctx, "purge", fs+p)
	return nil, err
}