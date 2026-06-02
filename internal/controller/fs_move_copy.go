package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"
	"rcloneflow/internal/logger"
)

func (c *FsController) HandleMove(w http.ResponseWriter, r *http.Request) { c.wrap(w, r, c.doMoveFile) }
func (c *FsController) HandleCopy(w http.ResponseWriter, r *http.Request) { c.wrap(w, r, c.doCopyFile) }
func (c *FsController) HandleCopyDir(w http.ResponseWriter, r *http.Request) {
	c.wrap(w, r, c.doCopyDir)
}
func (c *FsController) HandleMoveDir(w http.ResponseWriter, r *http.Request) {
	c.wrap(w, r, c.doMoveDir)
}

func (c *FsController) doCopyFile(ctx context.Context, body []byte) (any, error) {
	var req copyMoveFileReq
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("无效的请求格式: %w", err)
	}
	srcFs, src := normalize(sanitizeFsRemote(req.SrcFs), sanitizePath(req.SrcRemote))
	dstFs, dst := normalize(sanitizeFsRemote(req.DstFs), sanitizePath(req.DstRemote))
	ensureDir(ctx, dstFs, parentDir(dst))
	_, err := runRclone(ctx, "copyto", srcFs+src, dstFs+dst)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "network name not found") || strings.Contains(strings.ToLower(err.Error()), "create filesystem") {
			s2 := tryStripFirstSegment(src)
			d2 := tryStripFirstSegment(dst)
			if s2 != src || d2 != dst {
				ensureDir(ctx, dstFs, parentDir(d2))
				if _, err2 := runRclone(ctx, "copyto", srcFs+s2, dstFs+d2); err2 == nil {
					return nil, nil
				} else {
					err = err2
				}
			}
		}
	}
	return nil, err
}

func (c *FsController) doMoveFile(ctx context.Context, body []byte) (any, error) {
	var req copyMoveFileReq
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("无效的请求格式: %w", err)
	}
	srcFs, src := normalize(sanitizeFsRemote(req.SrcFs), sanitizePath(req.SrcRemote))
	dstFs, dst := normalize(sanitizeFsRemote(req.DstFs), sanitizePath(req.DstRemote))
	_, err := runRclone(ctx, "moveto", srcFs+src, dstFs+dst)
	if err == nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("goroutine panic", zap.Any("panic", r))
				}
			}()
			waitGoneFile(context.Background(), srcFs, src)
		}()
		return nil, nil
	}
	if strings.Contains(strings.ToLower(err.Error()), "network name not found") || strings.Contains(strings.ToLower(err.Error()), "create filesystem") {
		s2 := tryStripFirstSegment(src)
		d2 := tryStripFirstSegment(dst)
		if s2 != src || d2 != dst {
			if _, err2 := runRclone(ctx, "moveto", srcFs+s2, dstFs+d2); err2 == nil {
				go func() {
					defer func() {
						if r := recover(); r != nil {
							logger.Error("goroutine panic", zap.Any("panic", r))
						}
					}()
					waitGoneFile(context.Background(), srcFs, s2)
				}()
				return nil, nil
			} else {
				err = err2
			}
		}
	}
	if isWebdavMoveError(err) {
		if _, er2 := runRclone(ctx, "copyto", srcFs+src, dstFs+dst); er2 == nil {
			_, _ = runRclone(ctx, "deletefile", srcFs+src)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Error("goroutine panic", zap.Any("panic", r))
					}
				}()
				waitVisible(context.Background(), dstFs, dst)
			}()
			go func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Error("goroutine panic", zap.Any("panic", r))
					}
				}()
				waitGoneFile(context.Background(), srcFs, src)
			}()
			return nil, nil
		}
	}
	return nil, err
}

func (c *FsController) doCopyDir(ctx context.Context, body []byte) (any, error) {
	var req copyMoveFileReq
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("无效的请求格式: %w", err)
	}
	srcFs, src := splitFsRemote(sanitizeFsRemote(req.SrcFs), sanitizePath(req.SrcRemote))
	dstFs, dst := splitFsRemote(sanitizeFsRemote(req.DstFs), sanitizePath(req.DstRemote))
	ensureDir(ctx, dstFs, dst)
	_, err := runRclone(ctx, "copy", srcFs+src, dstFs+dst)
	if err != nil {
		low := strings.ToLower(err.Error())
		if strings.Contains(low, "network name not found") || strings.Contains(low, "create filesystem") {
			s2 := tryStripFirstSegment(src)
			d2 := tryStripFirstSegment(dst)
			if s2 != src || d2 != dst {
				ensureDir(ctx, dstFs, d2)
				if _, err2 := runRclone(ctx, "copy", srcFs+s2, dstFs+d2); err2 == nil {
					return nil, nil
				} else {
					err = err2
				}
			}
		}
	}
	return nil, err
}

func (c *FsController) doMoveDir(ctx context.Context, body []byte) (any, error) {
	var req copyMoveFileReq
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("无效的请求格式: %w", err)
	}
	srcFs, src := splitFsRemote(sanitizeFsRemote(req.SrcFs), sanitizePath(req.SrcRemote))
	dstFs, dst := splitFsRemote(sanitizeFsRemote(req.DstFs), sanitizePath(req.DstRemote))
	ensureDir(ctx, dstFs, dst)
	_, err := runRclone(ctx, "move", srcFs+src, dstFs+dst)
	if err == nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("goroutine panic", zap.Any("panic", r))
				}
			}()
			deepCleanDir(context.Background(), srcFs, src)
			waitGoneDir(context.Background(), srcFs, src)
		}()
		return nil, nil
	}
	if strings.Contains(strings.ToLower(err.Error()), "network name not found") || strings.Contains(strings.ToLower(err.Error()), "create filesystem") {
		s2 := tryStripFirstSegment(src)
		d2 := tryStripFirstSegment(dst)
		if s2 != src || d2 != dst {
			ensureDir(ctx, dstFs, d2)
			if _, err2 := runRclone(ctx, "move", srcFs+s2, dstFs+d2); err2 == nil {
				go func() {
					defer func() {
						if r := recover(); r != nil {
							logger.Error("goroutine panic", zap.Any("panic", r))
						}
					}()
					deepCleanDir(context.Background(), srcFs, s2)
					waitGoneDir(context.Background(), srcFs, s2)
				}()
				return nil, nil
			} else {
				err = err2
			}
		}
	}
	if isWebdavMoveError(err) {
		if _, er2 := runRclone(ctx, "copy", srcFs+src, dstFs+dst); er2 == nil {
			_, _ = runRclone(ctx, "purge", srcFs+src)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Error("goroutine panic", zap.Any("panic", r))
					}
				}()
				waitVisible(context.Background(), dstFs, dst)
			}()
			go func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Error("goroutine panic", zap.Any("panic", r))
					}
				}()
				deepCleanDir(context.Background(), srcFs, src)
				waitGoneDir(context.Background(), srcFs, src)
			}()
			return nil, nil
		}
	}
	return nil, err
}
