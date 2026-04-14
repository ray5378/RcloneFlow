package rclone

import (
	"context"
	"strings"

	"rcloneflow/internal/adapter"
)

// Client rclone API客户端
type Client struct {
	cli *adapter.RcloneClient
}

// NewFromEnv 创建客户端（从环境变量读取配置）
func NewFromEnv() *Client {
	return &Client{
		cli: adapter.NewRcloneClient(nil),
	}
}

// NewWithConfig 使用指定配置创建客户端
func NewWithConfig(cfg *adapter.RcloneConfig) *Client {
	return &Client{
		cli: adapter.NewRcloneClient(cfg),
	}
}

// ListPath 列出路径
func (c *Client) ListPath(ctx context.Context, fs, remote string) ([]map[string]any, error) {
	items, err := c.cli.ListPath(ctx, fs, remote)
	if err != nil {
		return nil, err
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
	return result, nil
}

// Version 获取版本
func (c *Client) Version(ctx context.Context) (string, error) {
	v, err := c.cli.Version(ctx)
	if err != nil {
		return "", err
	}
	return v.Version, nil
}

// ListRemotes 获取远程存储列表
func (c *Client) ListRemotes(ctx context.Context) ([]string, error) {
	return c.cli.ListRemotes(ctx)
}

// CreateRemote 创建远程存储
func (c *Client) CreateRemote(ctx context.Context, name, typ string, params map[string]any) error {
	return c.cli.CreateRemote(ctx, &adapter.CreateRemoteRequest{
		Name:       name,
		Type:       typ,
		Parameters: params,
	})
}

// GetConfig 获取配置
func (c *Client) GetConfig(ctx context.Context, name string) (map[string]any, error) {
	return c.cli.GetConfig(ctx, name)
}

// DeleteRemote 删除远程存储
func (c *Client) DeleteRemote(ctx context.Context, name string) error {
	return c.cli.DeleteRemote(ctx, name)
}

// DumpConfig 导出配置
func (c *Client) DumpConfig(ctx context.Context) (map[string]map[string]any, error) {
	return c.cli.DumpConfig(ctx)
}

// GetProviders 获取提供商
func (c *Client) GetProviders(ctx context.Context) ([]map[string]any, error) {
	providers, err := c.cli.GetProviders(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]any, len(providers))
	for i, p := range providers {
		result[i] = map[string]any{
			"Name":      p.Name,
			"Hangul":    p.Hangul,
			"Prefix":    p.Prefix,
			"OpenURL":   p.OpenURL,
			"HashTypes": p.HashTypes,
			"Options":   p.Options,
		}
	}
	return result, nil
}

// GetUsage 获取使用量
func (c *Client) GetUsage(ctx context.Context, fs string) (map[string]any, error) {
	about, err := c.cli.GetUsage(ctx, fs)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"used":  about.Used,
		"free":  about.Free,
		"total": about.Used + about.Free,
	}, nil
}

// GetFsInfo 获取文件系统信息
func (c *Client) GetFsInfo(ctx context.Context, fs string) (map[string]any, error) {
	info, err := c.cli.GetFsInfo(ctx, fs)
	if err != nil {
		return nil, err
	}
	result := map[string]any{
		"name":      info.Name,
		"precision": info.Precision,
		"root":      info.Root,
	}
	if info.Features != nil {
		result["features"] = info.Features
	}
	return result, nil
}

// Deprecated: The following RC-based FS helpers are retained only for diagnostics/fallback.
// Production FS operations are CLI-only in internal/controller/fs_cli.go and are not routed here.

// Mkdir 创建目录（Deprecated: not wired to API routes）
func (c *Client) Mkdir(ctx context.Context, fs, remote string) error {
	return c.cli.Mkdir(ctx, fs, remote)
}

// DeleteFile 删除文件（Deprecated: not wired to API routes）
func (c *Client) DeleteFile(ctx context.Context, fs, remote string) error {
	return c.cli.DeleteFile(ctx, fs, remote)
}

// Purge 删除目录（Deprecated: not wired to API routes）
func (c *Client) Purge(ctx context.Context, fs, remote string) error {
	return c.cli.Purge(ctx, fs, remote)
}

// MoveFile 移动文件（Deprecated: not wired to API routes）
func (c *Client) MoveFile(ctx context.Context, srcFs, srcRemote, dstFs, dstRemote string) error {
	return c.cli.MoveFile(ctx, srcFs, srcRemote, dstFs, dstRemote)
}

// CopyFile 复制文件（Deprecated: not wired to API routes）
func (c *Client) CopyFile(ctx context.Context, srcFs, srcRemote, dstFs, dstRemote string) error {
	return c.cli.CopyFile(ctx, srcFs, srcRemote, dstFs, dstRemote)
}

// CopyDir 复制目录（Deprecated: not wired to API routes）
func (c *Client) CopyDir(ctx context.Context, srcFs, dstFs string) error {
	return c.cli.CopyDir(ctx, srcFs, dstFs)
}

// MoveDir 移动目录（Deprecated: not wired to API routes）
func (c *Client) MoveDir(ctx context.Context, srcFs, dstFs string) error {
	return c.cli.MoveDir(ctx, srcFs, dstFs)
}

// PublicLink 生成分享链接（Deprecated: not wired to API routes）
func (c *Client) PublicLink(ctx context.Context, fs, remote string) (string, error) {
	return c.cli.PublicLink(ctx, fs, remote)
}

// StartJob 启动任务
func (c *Client) StartJob(ctx context.Context, mode, srcFs, dstFs string, opts *adapter.TaskOptions) (int64, error) {
	return c.cli.StartJob(ctx, mode, srcFs, dstFs, opts)
}

// JobStatus 获取任务状态
func (c *Client) JobStatus(ctx context.Context, jobID int64) (map[string]any, error) {
	status, err := c.cli.JobStatus(ctx, jobID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"finished": status.Finished,
		"success":  status.Success,
		"error":    status.Error,
	}, nil
}

// RunTask 运行任务
func (c *Client) RunTask(ctx context.Context, taskID int64, mode, srcRemote, srcPath, dstRemote, dstPath, trigger string, opts *adapter.TaskOptions) (int64, error) {
	src := srcRemote + ":" + strings.TrimPrefix(srcPath, "/")
	dst := dstRemote + ":" + strings.TrimPrefix(dstPath, "/")
	return c.cli.StartJob(ctx, mode, src, dst, opts)
}

// CoreStats 获取 rclone 核心统计信息
func (c *Client) CoreStats(ctx context.Context) (map[string]any, error) {
	var resp map[string]any
	err := c.cli.Call(ctx, "core/stats", nil, &resp)
	return resp, err
}

func (c *Client) CoreStatsGroup(ctx context.Context, group string) (map[string]any, error) {
	var resp map[string]any
	err := c.cli.Call(ctx, "core/stats", map[string]any{"group": group}, &resp)
	return resp, err
}

// JobStop 停止指定的 Job
func (c *Client) JobStop(ctx context.Context, jobID int64) error {
	params := map[string]any{"jobid": jobID}
	var resp map[string]any
	return c.cli.Call(ctx, "job/stop", params, &resp)
}
