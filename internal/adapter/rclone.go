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

// RcloneConfig rclone连接配置
type RcloneConfig struct {
	BaseURL string        // rclone RC地址，如 http://127.0.0.1:5572
	User    string        // 用户名
	Pass    string        // 密码
	Timeout time.Duration // 请求超时时间
}

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

// RcloneClient rclone API客户端
type RcloneClient struct {
	config *RcloneConfig
	client *http.Client
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
		// 空请求也发送空JSON对象，避免 rclone EOF 错误
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

// VersionResponse 版本信息响应
type VersionResponse struct {
	Version    string `json:"version"`
	Decomposed []int  `json:"decomposed"`
	IsGit     bool   `json:"isGit"`
	IsBeta    bool   `json:"isBeta"`
	Os        string `json:"os"`
	OsKernel  string `json:"osKernel"`
	OsVersion string `json:"osVersion"`
	OsArch    string `json:"osArch"`
	Arch      string `json:"arch"`
	GoVersion string `json:"goVersion"`
	Linking   string `json:"linking"`
	GoTags    string `json:"goTags"`
}

// Version 获取rclone版本
func (c *RcloneClient) Version(ctx context.Context) (*VersionResponse, error) {
	var resp VersionResponse
	if err := c.Call(ctx, "core/version", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ==================== 远程存储配置API ====================

// ListRemotesResponse 远程存储列表响应
type ListRemotesResponse struct {
	Remotes []string `json:"remotes"`
}

// ListRemotes 获取所有远程存储名称
func (c *RcloneClient) ListRemotes(ctx context.Context) ([]string, error) {
	var resp ListRemotesResponse
	if err := c.Call(ctx, "config/listremotes", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Remotes, nil
}

// CreateRemoteRequest 创建远程存储请求
type CreateRemoteRequest struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Parameters map[string]any    `json:"parameters"`
	Opt        *CreateRemoteOpt  `json:"opt,omitempty"`
}

// CreateRemoteOpt 创建远程存储选项
type CreateRemoteOpt struct {
	Obscure   bool `json:"obscure"`
	NoOutput  bool `json:"noOutput"`
	NoPrompts bool `json:"noPrompts"`
}

// CreateRemote 创建远程存储
func (c *RcloneClient) CreateRemote(ctx context.Context, req *CreateRemoteRequest) error {
	// 设置默认选项
	if req.Opt == nil {
		req.Opt = &CreateRemoteOpt{Obscure: false, NoOutput: true}
	}
	
	// 清理参数
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
		// encoding字段使用默认值
		if k == "encoding" {
			if s, ok := v.(string); ok && s != "" {
				cleanParams[k] = s
			} else {
				cleanParams[k] = "Slash,LtGt,DoubleQuote,Colon,Question,Asterisk,Pipe,BackSlash,Ctl,RightSpace,RightPeriod,InvalidUtf8,Dot"
			}
			continue
		}
		// headers字段不传
		if k == "headers" {
			continue
		}
		cleanParams[k] = v
	}
	req.Parameters = cleanParams
	
	return c.Call(ctx, "config/create", req, nil)
}

// DeleteRemoteRequest 删除远程存储请求
type DeleteRemoteRequest struct {
	Name string `json:"name"`
}

// DeleteRemote 删除远程存储
func (c *RcloneClient) DeleteRemote(ctx context.Context, name string) error {
	return c.Call(ctx, "config/delete", &DeleteRemoteRequest{Name: name}, nil)
}

// DumpConfigResponse 配置导出响应
type DumpConfigResponse map[string]map[string]any

// DumpConfig 导出所有配置
func (c *RcloneClient) DumpConfig(ctx context.Context) (DumpConfigResponse, error) {
	var resp DumpConfigResponse
	if err := c.Call(ctx, "config/dump", nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetConfigRequest 获取单个配置请求
type GetConfigRequest struct {
	Name string `json:"name"`
}

// GetConfig 获取单个远程存储配置
func (c *RcloneClient) GetConfig(ctx context.Context, name string) (map[string]any, error) {
	var resp map[string]any
	if err := c.Call(ctx, "config/get", &GetConfigRequest{Name: name}, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// ProviderInfo 提供商信息
type ProviderInfo struct {
	Name     string         `json:"Name"`
	Hangul   string         `json:"Hangul,omitempty"`
	Prefix   string         `json:"Prefix,omitempty"`
	OpenURL  string         `json:"OpenURL,omitempty"`
	HashTypes []string      `json:"HashTypes,omitempty"`
	Attr     []string       `json:"Attr,omitempty"`
	Policy   []string       `json:"Policy,omitempty"`
	Options  []OptionBlock  `json:"Options,omitempty"`
}

// OptionBlock 选项块
type OptionBlock struct {
	Name       string      `json:"Name"`
	FieldName  string      `json:"FieldName,omitempty"`
	Help       string      `json:"Help"`
	Groups     string      `json:"Groups,omitempty"`
	Provider   string      `json:"Provider,omitempty"`
	Default    any         `json:"Default"`
	Value      any         `json:"Value,omitempty"`
	DefaultStr string      `json:"DefaultStr,omitempty"`
	ValueStr   string      `json:"ValueStr,omitempty"`
	Examples   []Example   `json:"Examples,omitempty"`
	ShortOpt   string      `json:"ShortOpt,omitempty"`
	Hide       int         `json:"Hide"`
	Required   bool        `json:"Required"`
	IsPassword bool        `json:"IsPassword"`
	NoPrefix   bool        `json:"NoPrefix"`
	Advanced   bool        `json:"Advanced"`
	Exclusive  bool        `json:"Exclusive"`
	Sensitive  bool        `json:"Sensitive"`
}

// Example 示例
type Example struct {
	Help  string `json:"Help"`
	Value any    `json:"Value"`
}

// ProvidersResponse 提供商列表响应
type ProvidersResponse struct {
	Providers []ProviderInfo `json:"providers"`
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

// PathInfo 路径信息
type PathInfo struct {
	Name      string         `json:"Name"`
	Path      string         `json:"Path"`
	IsDir     bool           `json:"IsDir"`
	IsBucket  bool           `json:"IsBucket,omitempty"`
	MimeType  string         `json:"MimeType,omitempty"`
	ModTime   string         `json:"ModTime,omitempty"`
	Size      int64          `json:"Size"`
	Hash      map[string]any `json:"Hash,omitempty"`
}

// ListPathRequest 列出路径请求
type ListPathRequest struct {
	Fs         string `json:"fs"`
	Remote     string `json:"remote"`
	Opt        *ListOpt `json:"opt,omitempty"`
}

// ListOpt 列出选项
type ListOpt struct {
	Recurse       bool     `json:"recurse"`
	NoModTime     bool     `json:"noModTime"`
	DirsOnly      bool     `json:"dirsOnly"`
	FilesOnly     bool     `json:"filesOnly"`
	ShowHash      []string `json:"showHash,omitempty"`
	IgnoreSize    bool     `json:"ignoreSize"`
	IncludeEmpty  bool     `json:"includeEmpty"`
	Flatten       []int    `json:"flatten,omitempty"`
}

// ListPathResponse 列出路径响应
type ListPathResponse struct {
	Fs   string     `json:"fs"`
	List []PathInfo `json:"list"`
}

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

// MkdirRequest 创建目录请求
type MkdirRequest struct {
	Fs     string `json:"fs"`
	Remote string `json:"remote"`
}

// Mkdir 创建目录
func (c *RcloneClient) Mkdir(ctx context.Context, fs, remote string) error {
	return c.Call(ctx, "operations/mkdir", &MkdirRequest{Fs: fs, Remote: remote}, nil)
}

// DeleteFileRequest 删除文件请求
type DeleteFileRequest struct {
	Fs     string `json:"fs"`
	Remote string `json:"remote"`
}

// DeleteFile 删除文件
func (c *RcloneClient) DeleteFile(ctx context.Context, fs, remote string) error {
	return c.Call(ctx, "operations/deletefile", &DeleteFileRequest{Fs: fs, Remote: remote}, nil)
}

// PurgeRequest 清空目录请求
type PurgeRequest struct {
	Fs     string `json:"fs"`
	Remote string `json:"remote"`
}

// Purge 删除目录及所有内容
func (c *RcloneClient) Purge(ctx context.Context, fs, remote string) error {
	return c.Call(ctx, "operations/purge", &PurgeRequest{Fs: fs, Remote: remote}, nil)
}

// MoveFileRequest 移动文件请求
type MoveFileRequest struct {
	SrcFs     string `json:"srcFs"`
	SrcRemote string `json:"srcRemote"`
	DstFs     string `json:"dstFs"`
	DstRemote string `json:"dstRemote"`
}

// MoveFile 移动文件
func (c *RcloneClient) MoveFile(ctx context.Context, srcFs, srcRemote, dstFs, dstRemote string) error {
	return c.Call(ctx, "operations/movefile", &MoveFileRequest{
		SrcFs: srcFs, SrcRemote: srcRemote,
		DstFs: dstFs, DstRemote: dstRemote,
	}, nil)
}

// CopyFileRequest 复制文件请求
type CopyFileRequest struct {
	SrcFs     string `json:"srcFs"`
	SrcRemote string `json:"srcRemote"`
	DstFs     string `json:"dstFs"`
	DstRemote string `json:"dstRemote"`
	Async     bool   `json:"_async"`
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
	// 等待异步任务完成
	return c.waitForJob(ctx, result.JobID)
}

// AboutResponse 存储使用量响应
type AboutResponse struct {
	Used   int64  `json:"used"`
	Trashed int64 `json:"trashed,omitempty"`
	Other   int64 `json:"other,omitempty"`
	Free    int64 `json:"free"`
}

// GetUsage 获取存储使用量
func (c *RcloneClient) GetUsage(ctx context.Context, fs string) (*AboutResponse, error) {
	var resp AboutResponse
	if err := c.Call(ctx, "operations/about", &FsRequest{Fs: fs}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// FsRequest 文件系统请求
type FsRequest struct {
	Fs string `json:"fs"`
}

// FsInfoResponse 文件系统信息响应
type FsInfoResponse struct {
	Name      string           `json:"Name"`
	Precision int64            `json:"Precision"`
	Root      string           `json:"Root"`
	String    string           `json:"String"`
	Features  *Features        `json:"Features,omitempty"`
	Hashes    []string         `json:"Hashes,omitempty"`
	MetadataInfo *MetadataInfo `json:"MetadataInfo,omitempty"`
}

// Features 功能特性
type Features struct {
	About             bool `json:"About"`
	BucketBased       bool `json:"BucketBased"`
	BucketBasedRootOK bool `json:"BucketBasedRootOK"`
	CanHaveEmptyDirectories bool `json:"CanHaveEmptyDirectories"`
	CaseInsensitive   bool `json:"CaseInsensitive"`
	ChangeNotify      bool `json:"ChangeNotify"`
	CleanUp           bool `json:"CleanUp"`
	Command           bool `json:"Command"`
	Copy              bool `json:"Copy"`
	DirCacheFlush     bool `json:"DirCacheFlush"`
	DirMove           bool `json:"DirMove"`
	Disconnect        bool `json:"Disconnect"`
	DuplicateFiles    bool `json:"DuplicateFiles"`
	GetTier           bool `json:"GetTier"`
	IsLocal           bool `json:"IsLocal"`
	ListR             bool `json:"ListR"`
	MergeDirs         bool `json:"MergeDirs"`
	MetadataInfo      bool `json:"MetadataInfo"`
	Move              bool `json:"Move"`
	OpenWriterAt      bool `json:"OpenWriterAt"`
	PublicLink        bool `json:"PublicLink"`
	Purge             bool `json:"Purge"`
	PutStream         bool `json:"PutStream"`
	PutUnchecked      bool `json:"PutUnchecked"`
	ReadMetadata      bool `json:"ReadMetadata"`
	ReadMimeType      bool `json:"ReadMimeType"`
	ServerSideAcrossConfigs bool `json:"ServerSideAcrossConfigs"`
	SetTier           bool `json:"SetTier"`
	SetWrapper        bool `json:"SetWrapper"`
	Shutdown          bool `json:"Shutdown"`
	SlowHash          bool `json:"SlowHash"`
	SlowModTime       bool `json:"SlowModTime"`
	UnWrap            bool `json:"UnWrap"`
	UserInfo          bool `json:"UserInfo"`
	UserMetadata      bool `json:"UserMetadata"`
	WrapFs            bool `json:"WrapFs"`
	WriteMetadata     bool `json:"WriteMetadata"`
	WriteMimeType     bool `json:"WriteMimeType"`
}

// MetadataInfo 元数据信息
type MetadataInfo struct {
	System map[string]MetadataField `json:"System,omitempty"`
}

// MetadataField 元数据字段
type MetadataField struct {
	Help    string `json:"Help"`
	Type    string `json:"Type"`
	Example string `json:"Example,omitempty"`
}

// GetFsInfo 获取文件系统信息
func (c *RcloneClient) GetFsInfo(ctx context.Context, fs string) (*FsInfoResponse, error) {
	var resp FsInfoResponse
	if err := c.Call(ctx, "operations/fsinfo", &FsRequest{Fs: fs}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// PublicLinkResponse 分享链接响应
type PublicLinkResponse struct {
	URL string `json:"url"`
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

// PublicLinkRequest 分享链接请求
type PublicLinkRequest struct {
	Fs           string `json:"fs"`
	Remote       string `json:"remote"`
	Expire       string `json:"expire,omitempty"`
	Unpublished  bool   `json:"unpublished"`
}

// ==================== 同步操作API ====================

// SyncCopyRequest 同步复制请求
type SyncCopyRequest struct {
	SrcFs               string `json:"srcFs"`
	DstFs               string `json:"dstFs"`
	CreateEmptySrcDirs bool   `json:"createEmptySrcDirs"`
}

// CopyDir 复制目录
func (c *RcloneClient) CopyDir(ctx context.Context, srcFs, dstFs string) error {
	return c.Call(ctx, "sync/copy", &SyncCopyRequest{
		SrcFs: srcFs, DstFs: dstFs,
		CreateEmptySrcDirs: true,
	}, nil)
}

// SyncMoveRequest 同步移动请求
type SyncMoveRequest struct {
	SrcFs               string `json:"srcFs"`
	DstFs               string `json:"dstFs"`
	CreateEmptySrcDirs bool   `json:"createEmptySrcDirs"`
	DeleteEmptySrcDirs bool   `json:"deleteEmptySrcDirs"`
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

// JobIDResponse 任务ID响应
type JobIDResponse struct {
	JobID int64 `json:"jobid"`
}

// StartJobRequest 启动任务请求
type StartJobRequest struct {
	SrcFs  string `json:"srcFs"`
	DstFs  string `json:"dstFs"`
	Async  bool   `json:"_async"`
	*TaskOptions
}

// StartJob 启动同步任务
func (c *RcloneClient) StartJob(ctx context.Context, mode, srcFs, dstFs string, opts *TaskOptions) (int64, error) {
	// sync/copy 或 sync/sync 或 sync/move
	ep := "sync/copy"
	switch strings.ToLower(mode) {
	case "sync":
		ep = "sync/sync"
	case "move":
		ep = "sync/move"
	}
	
	var resp JobIDResponse
	req := &StartJobRequest{
		SrcFs:        srcFs,
		DstFs:        dstFs,
		Async:        true,
		TaskOptions:  opts,
	}
	
	if err := c.Call(ctx, ep, req, &resp); err != nil {
		return 0, err
	}
	return resp.JobID, nil
}

// JobStatusResponse 任务状态响应
type JobStatusResponse struct {
	ID        int64             `json:"id"`
	ExecuteID string            `json:"executeId"`
	StartTime string            `json:"startTime"`
	EndTime   string            `json:"endTime,omitempty"`
	Duration  float64          `json:"duration"`
	Success   bool              `json:"success"`
	Finished  bool              `json:"finished"`
	Error     string            `json:"error,omitempty"`
	Output    map[string]any    `json:"output,omitempty"`
	Progress  map[string]any    `json:"progress,omitempty"`
}

// JobStatus 获取任务状态
func (c *RcloneClient) JobStatus(ctx context.Context, jobID int64) (*JobStatusResponse, error) {
	var resp JobStatusResponse
	if err := c.Call(ctx, "job/status", &JobIDRequest{JobID: jobID}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// JobIDRequest 任务ID请求
type JobIDRequest struct {
	JobID int64 `json:"jobid"`
}

// waitForJob 等待异步任务完成
func (c *RcloneClient) waitForJob(ctx context.Context, jobID int64) error {
	for {
		status, err := c.JobStatus(ctx, jobID)
		if err != nil {
			return err
		}
		if status.Finished {
			if !status.Success && status.Error != "" {
				return fmt.Errorf("job failed: %s", status.Error)
			}
			return nil
		}
		// 等待1秒
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

// JobListResponse 任务列表响应
type JobListResponse struct {
	ExecuteID   string  `json:"executeId"`
	JobIDs      []int64 `json:"jobids"`
	RunningIDs  []int64 `json:"running_ids"`
	FinishedIDs []int64 `json:"finished_ids"`
}

// ListJobs 列出所有任务
func (c *RcloneClient) ListJobs(ctx context.Context) (*JobListResponse, error) {
	var resp JobListResponse
	if err := c.Call(ctx, "job/list", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// StopJobRequest 停止任务请求
type StopJobRequest struct {
	JobID int64 `json:"jobid"`
}

// StopJob 停止任务
func (c *RcloneClient) StopJob(ctx context.Context, jobID int64) error {
	return c.Call(ctx, "job/stop", &StopJobRequest{JobID: jobID}, nil)
}
