package rclone

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

type Client struct {
    BaseURL string
    User    string
    Pass    string
    Client  *http.Client
}

func NewFromEnv() *Client {
    base := os.Getenv("RCLONE_RC_URL")
    if base == "" {
        base = "http://127.0.0.1:5572"
    }
    timeout := 120 * time.Second
    if v := strings.TrimSpace(os.Getenv("RCLONE_RC_TIMEOUT")); v != "" {
        if d, err := time.ParseDuration(v); err == nil && d > 0 {
            timeout = d
        }
    }
    return &Client{
        BaseURL: strings.TrimRight(base, "/"),
        User:    os.Getenv("RCLONE_RC_USER"),
        Pass:    os.Getenv("RCLONE_RC_PASS"),
        Client:  &http.Client{Timeout: timeout},
    }
}

func (c *Client) Call(ctx context.Context, endpoint string, req any, out any) error {
    if req == nil {
        req = map[string]any{}
    }
    bs, err := json.Marshal(req)
    if err != nil {
        return err
    }
    httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/"+strings.TrimLeft(endpoint, "/"), bytes.NewReader(bs))
    if err != nil {
        return err
    }
    httpReq.Header.Set("Content-Type", "application/json")
    if c.User != "" || c.Pass != "" {
        httpReq.SetBasicAuth(c.User, c.Pass)
    }
    resp, err := c.Client.Do(httpReq)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    if resp.StatusCode >= 300 {
        return fmt.Errorf("rclone rc %s failed: %s", endpoint, strings.TrimSpace(string(body)))
    }
    if out != nil {
        if len(body) == 0 {
            return nil
        }
        return json.Unmarshal(body, out)
    }
    return nil
}

func (c *Client) Version(ctx context.Context) (string, error) {
    var out map[string]any
    if err := c.Call(ctx, "core/version", nil, &out); err != nil {
        return "", err
    }
    if v, ok := out["version"].(string); ok {
        return v, nil
    }
    return "unknown", nil
}

func (c *Client) ListRemotes(ctx context.Context) ([]string, error) {
    var out struct {
        Remotes []string `json:"remotes"`
    }
    if err := c.Call(ctx, "config/listremotes", nil, &out); err != nil {
        return nil, err
    }
    return out.Remotes, nil
}

func (c *Client) CreateRemote(ctx context.Context, name, typ string, params map[string]any) error {
    // 过滤可能导致问题的参数
    cleanParams := map[string]any{}
    for k, v := range params {
        // 跳过空数组和空字符串
        if v == nil { continue }
        if s, ok := v.(string); ok && s == "" { continue }
        if arr, ok := v.([]any); ok && len(arr) == 0 { continue }
        
        // encoding 字段使用 API 返回的正确值
        if k == "encoding" {
            if s, ok := v.(string); ok && s != "" {
                cleanParams[k] = s
            } else {
                cleanParams[k] = "Slash,LtGt,DoubleQuote,Colon,Question,Asterisk,Pipe,BackSlash,Ctl,RightSpace,RightPeriod,InvalidUtf8,Dot"
            }
            continue
        }
        
        // headers 字段不传
        if k == "headers" { continue }
        
        cleanParams[k] = v
    }
    
    req := map[string]any{"name": name, "type": typ, "parameters": cleanParams, "opt": map[string]any{"obscure": false, "noOutput": false}}
    return c.Call(ctx, "config/create", req, nil)
}

func (c *Client) ListPath(ctx context.Context, fs, remote string) ([]map[string]any, error) {
    var out struct {
        List []map[string]any `json:"list"`
    }
    if err := c.Call(ctx, "operations/list", map[string]any{"fs": fs, "remote": strings.TrimPrefix(remote, "/")}, &out); err != nil {
        return nil, err
    }
    return out.List, nil
}

func (c *Client) StartJob(ctx context.Context, mode, srcFs, dstFs string) (int64, error) {
    ep := "sync/copy"
    if strings.ToLower(mode) == "sync" {
        ep = "sync/sync"
    }
    var out struct {
        JobID int64 `json:"jobid"`
    }
    err := c.Call(ctx, ep, map[string]any{"srcFs": srcFs, "dstFs": dstFs, "_async": true}, &out)
    return out.JobID, err
}

func (c *Client) JobStatus(ctx context.Context, jobID int64) (map[string]any, error) {
    var out map[string]any
    err := c.Call(ctx, "job/status", map[string]any{"jobid": jobID}, &out)
    return out, err
}

// GetProviders 获取所有支持的存储类型及其配置选项
func (c *Client) GetProviders(ctx context.Context) ([]map[string]any, error) {
    var out struct {
        Providers []map[string]any `json:"providers"`
    }
    if err := c.Call(ctx, "config/providers", nil, &out); err != nil {
        return nil, err
    }
    return out.Providers, nil
}

// DumpConfig 获取所有存储配置
func (c *Client) DumpConfig(ctx context.Context) (map[string]map[string]any, error) {
    var out map[string]map[string]any
    if err := c.Call(ctx, "config/dump", nil, &out); err != nil {
        return nil, err
    }
    return out, nil
}

// GetConfig 获取单个存储配置
func (c *Client) GetConfig(ctx context.Context, name string) (map[string]any, error) {
    var out map[string]any
    if err := c.Call(ctx, "config/get", map[string]any{"name": name}, &out); err != nil {
        return nil, err
    }
    return out, nil
}

// DeleteRemote 删除存储
func (c *Client) DeleteRemote(ctx context.Context, name string) error {
    return c.Call(ctx, "config/delete", map[string]any{"name": name}, nil)
}

// GetUsage 获取存储使用量
func (c *Client) GetUsage(ctx context.Context, fs string) (map[string]any, error) {
    var out map[string]any
    if err := c.Call(ctx, "operations/about", map[string]any{"fs": fs}, &out); err != nil {
        return nil, err
    }
    return out, nil
}

// GetFsInfo 获取文件系统信息
func (c *Client) GetFsInfo(ctx context.Context, fs string) (map[string]any, error) {
    var out map[string]any
    if err := c.Call(ctx, "operations/fsinfo", map[string]any{"fs": fs}, &out); err != nil {
        return nil, err
    }
    return out, nil
}

// Mkdir 创建目录
func (c *Client) Mkdir(ctx context.Context, fs, remote string) error {
    return c.Call(ctx, "operations/mkdir", map[string]any{"fs": fs, "remote": remote}, nil)
}

// DeleteFile 删除文件
func (c *Client) DeleteFile(ctx context.Context, srcFs, srcRemote string) error {
    return c.Call(ctx, "operations/deletefile", map[string]any{"fs": srcFs, "remote": srcRemote}, nil)
}

// Purge 删除目录
func (c *Client) Purge(ctx context.Context, fs, remote string) error {
    return c.Call(ctx, "operations/purge", map[string]any{"fs": fs, "remote": remote}, nil)
}

// MoveFile 移动/重命名文件
func (c *Client) MoveFile(ctx context.Context, srcFs, srcRemote, dstFs, dstRemote string) error {
    return c.Call(ctx, "operations/movefile", map[string]any{
        "srcFs": srcFs, "srcRemote": srcRemote,
        "dstFs": dstFs, "dstRemote": dstRemote,
    }, nil)
}

// CopyFile 复制文件
func (c *Client) CopyFile(ctx context.Context, srcFs, srcRemote, dstFs, dstRemote string) error {
    return c.Call(ctx, "operations/copyfile", map[string]any{
        "srcFs": srcFs, "srcRemote": srcRemote,
        "dstFs": dstFs, "dstRemote": dstRemote,
    }, nil)
}

// PublicLink 生成分享链接
func (c *Client) PublicLink(ctx context.Context, fs, remote string) (string, error) {
    var out struct {
        URL string `json:"url"`
    }
    if err := c.Call(ctx, "operations/publiclink", map[string]any{"fs": fs, "remote": remote}, &out); err != nil {
        return "", err
    }
    return out.URL, nil
}
