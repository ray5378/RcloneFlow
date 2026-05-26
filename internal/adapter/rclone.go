package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// DefaultConfig 返回默认配置
func DefaultConfig() *RcloneConfig {
	cfg := &RcloneConfig{
		BaseURL: "http://127.0.0.1:5572",
		Timeout: 120 * time.Second,
	}
	if v := os.Getenv("RCLONE_RC_URL"); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv("RCLONE_RC_USER"); v != "" {
		cfg.User = v
	}
	if v := os.Getenv("RCLONE_RC_PASS"); v != "" {
		cfg.Pass = v
	}
	if v := strings.TrimSpace(os.Getenv("RCLONE_RC_TIMEOUT")); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			cfg.Timeout = d
		}
	}
	return cfg
}

// NewRcloneClient 创建rclone客户端
func NewRcloneClient(cfg *RcloneConfig) *RcloneClient {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &RcloneClient{
		config: cfg,
		client: &http.Client{Timeout: cfg.Timeout},
	}
}

// Call 调用rclone API
func (c *RcloneClient) Call(ctx context.Context, endpoint string, req, resp interface{}) error {
	var body []byte
	var err error

	if req != nil {
		body, err = json.Marshal(req)
		if err != nil {
			return fmt.Errorf("marshal request failed: %w", err)
		}
	} else {
		body = []byte("{}")
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(c.config.BaseURL, "/")+"/"+strings.TrimLeft(endpoint, "/"),
		bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.config.User != "" || c.config.Pass != "" {
		httpReq.SetBasicAuth(c.config.User, c.config.Pass)
	}

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	if httpResp == nil {
		return fmt.Errorf("request failed: nil response")
	}
	defer httpResp.Body.Close()

	respBody, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode >= 300 {
		return fmt.Errorf("rclone %s failed: %s", endpoint, strings.TrimSpace(string(respBody)))
	}

	if resp != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, resp); err != nil {
			return fmt.Errorf("unmarshal response failed: %w", err)
		}
	}
	return nil
}

// ==================== 核心API ====================

// Version 获取rclone版本
func (c *RcloneClient) Version(ctx context.Context) (*VersionResponse, error) {
	var resp VersionResponse
	if err := c.Call(ctx, "core/version", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ==================== 远程存储配置API ====================

// ListRemotes 获取所有远程存储名称
func (c *RcloneClient) ListRemotes(ctx context.Context) ([]string, error) {
	var resp ListRemotesResponse
	if err := c.Call(ctx, "config/listremotes", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Remotes, nil
}

// CreateRemote 创建远程存储
func (c *RcloneClient) CreateRemote(ctx context.Context, req *CreateRemoteRequest) error {
	if req.Opt == nil {
		req.Opt = &CreateRemoteOpt{Obscure: false, NoOutput: true}
	}

	cleanParams := make(map[string]any)
	for k, v := range req.Parameters {
		if v == nil {
			continue
		}
		if s, ok := v.(string); ok && s == "" {
			continue
		}
		if arr, ok := v.([]any); ok && len(arr) == 0 {
			continue
		}
		if k == "encoding" {
			if s, ok := v.(string); ok && s != "" {
				cleanParams[k] = s
			} else {
				cleanParams[k] = "Slash,LtGt,DoubleQuote,Colon,Question,Asterisk,Pipe,BackSlash,Ctl,RightSpace,RightPeriod,InvalidUtf8,Dot"
			}
			continue
		}
		if k == "headers" {
			continue
		}
		cleanParams[k] = v
	}
	req.Parameters = cleanParams

	return c.Call(ctx, "config/create", req, nil)
}

// DeleteRemote 删除远程存储
func (c *RcloneClient) DeleteRemote(ctx context.Context, name string) error {
	return c.Call(ctx, "config/delete", &DeleteRemoteRequest{Name: name}, nil)
}

// DumpConfig 导出所有配置
func (c *RcloneClient) DumpConfig(ctx context.Context) (DumpConfigResponse, error) {
	var resp DumpConfigResponse
	if err := c.Call(ctx, "config/dump", nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetConfig 获取单个远程存储配置
func (c *RcloneClient) GetConfig(ctx context.Context, name string) (map[string]any, error) {
	var resp map[string]any
	if err := c.Call(ctx, "config/get", &GetConfigRequest{Name: name}, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetProviders 获取所有存储提供商
func (c *RcloneClient) GetProviders(ctx context.Context) ([]ProviderInfo, error) {
	var resp ProvidersResponse
	if err := c.Call(ctx, "config/providers", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Providers, nil
}

// ==================== 文件系统操作API ====================

// ListPath 列出目录内容
func (c *RcloneClient) ListPath(ctx context.Context, fs, remote string) ([]PathInfo, error) {
	var resp ListPathResponse
	if err := c.Call(ctx, "operations/list", &ListPathRequest{
		Fs:     fs,
		Remote: strings.TrimPrefix(remote, "/"),
	}, &resp); err != nil {
		return nil, err
	}
	return resp.List, nil
}

// Mkdir 创建目录
func (c *RcloneClient) Mkdir(ctx context.Context, fs, remote string) error {
	return c.Call(ctx, "operations/mkdir", &MkdirRequest{Fs: fs, Remote: remote}, nil)
}

// DeleteFile 删除文件
func (c *RcloneClient) DeleteFile(ctx context.Context, fs, remote string) error {
	return c.Call(ctx, "operations/deletefile", &DeleteFileRequest{Fs: fs, Remote: remote}, nil)
}

// Purge 删除目录及所有内容
func (c *RcloneClient) Purge(ctx context.Context, fs, remote string) error {
	return c.Call(ctx, "operations/purge", &PurgeRequest{Fs: fs, Remote: remote}, nil)
}

// MoveFile 移动文件
func (c *RcloneClient) MoveFile(ctx context.Context, srcFs, srcRemote, dstFs, dstRemote string) error {
	return c.Call(ctx, "operations/movefile", &MoveFileRequest{
		SrcFs: srcFs, SrcRemote: srcRemote,
		DstFs: dstFs, DstRemote: dstRemote,
	}, nil)
}

// CopyFile 复制文件
func (c *RcloneClient) CopyFile(ctx context.Context, srcFs, srcRemote, dstFs, dstRemote string) error {
	var result JobIDResponse
	if err := c.Call(ctx, "operations/copyfile", &CopyFileRequest{
		SrcFs: srcFs, SrcRemote: srcRemote,
		DstFs: dstFs, DstRemote: dstRemote,
		Async: true,
	}, &result); err != nil {
		return err
	}
	return c.waitForJob(ctx, result.JobID)
}

// GetUsage 获取存储使用量
func (c *RcloneClient) GetUsage(ctx context.Context, fs string) (*AboutResponse, error) {
	var resp AboutResponse
	if err := c.Call(ctx, "operations/about", &FsRequest{Fs: fs}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetFsInfo 获取文件系统信息
func (c *RcloneClient) GetFsInfo(ctx context.Context, fs string) (*FsInfoResponse, error) {
	var resp FsInfoResponse
	if err := c.Call(ctx, "operations/fsinfo", &FsRequest{Fs: fs}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// PublicLink 生成分享链接
func (c *RcloneClient) PublicLink(ctx context.Context, fs, remote string) (string, error) {
	var resp PublicLinkResponse
	if err := c.Call(ctx, "operations/publiclink", &PublicLinkRequest{
		Fs: fs, Remote: remote,
	}, &resp); err != nil {
		return "", err
	}
	return resp.URL, nil
}

// ==================== 同步操作API ====================

// CopyDir 复制目录
func (c *RcloneClient) CopyDir(ctx context.Context, srcFs, dstFs string) error {
	return c.Call(ctx, "sync/copy", &SyncCopyRequest{
		SrcFs: srcFs, DstFs: dstFs,
		CreateEmptySrcDirs: true,
	}, nil)
}

// MoveDir 移动目录
func (c *RcloneClient) MoveDir(ctx context.Context, srcFs, dstFs string) error {
	return c.Call(ctx, "sync/move", &SyncMoveRequest{
		SrcFs: srcFs, DstFs: dstFs,
		CreateEmptySrcDirs: true,
		DeleteEmptySrcDirs: true,
	}, nil)
}

// ==================== 任务操作API ====================

// StartJob 启动同步任务
func (c *RcloneClient) StartJob(ctx context.Context, mode, srcFs, dstFs string, opts *TaskOptions) (int64, error) {
	ep := "sync/copy"
	switch strings.ToLower(mode) {
	case "sync":
		ep = "sync/sync"
	case "move":
		ep = "sync/move"
	}

	var resp JobIDResponse
	req := &StartJobRequest{
		SrcFs:       srcFs,
		DstFs:       dstFs,
		Async:       true,
		TaskOptions: opts,
	}

	if err := c.Call(ctx, ep, req, &resp); err != nil {
		return 0, err
	}
	return resp.JobID, nil
}

func (c *RcloneClient) jobStatus(ctx context.Context, jobID int64) (*jobStatusResponse, error) {
	var resp jobStatusResponse
	if err := c.Call(ctx, "job/status", &JobIDRequest{JobID: jobID}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// waitForJob 等待异步任务完成
func (c *RcloneClient) waitForJob(ctx context.Context, jobID int64) error {
	for {
		status, err := c.jobStatus(ctx, jobID)
		if err != nil {
			return err
		}
		if status.Finished {
			if !status.Success && status.Error != "" {
				return fmt.Errorf("job failed: %s", status.Error)
			}
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
}