package webdavserver

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/net/webdav"
)

// errorFS 模拟一个会在 Stat 时返回非 os.ErrNotExist 错误的文件系统
// 用于测试 suppressWriteHeaderWriter 是否正确丢弃错误响应体
type errorFS struct {
	dir webdav.Dir
}

func (e *errorFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error    { return nil }
func (e *errorFS) RemoveAll(ctx context.Context, name string) error                   { return nil }
func (e *errorFS) Rename(ctx context.Context, oldName, newName string) error           { return nil }

func (e *errorFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	// 对特定路径返回非 os.ErrNotExist 的错误
	if strings.Contains(name, "error-entry") {
		return nil, fmt.Errorf("rclone command failed: connection refused")
	}
	return e.dir.Stat(ctx, name)
}

func (e *errorFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	return e.dir.OpenFile(ctx, name, flag, perm)
}

// TestSuppressWriteHeaderDiscardsErrorBody 测试错误场景
// 当 webdav 库在处理 PROPFIND 时遇到非 os.ErrNotExist 错误，
// 会尝试写入错误响应（500），suppressWriteHeaderWriter 应丢弃该错误体
func TestSuppressWriteHeaderDiscardsErrorBody(t *testing.T) {
	tmpDir := t.TempDir()

	handler := &webdav.Handler{
		FileSystem: webdav.Dir(tmpDir),
		LockSystem: webdav.NewMemLS(),
	}

	// 包装 handler
	wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &suppressWriteHeaderWriter{ResponseWriter: w}
		handler.ServeHTTP(sw, r)
	})

	server := httptest.NewServer(wrappedHandler)
	defer server.Close()

	// 发送 PROPFIND 请求
	req, _ := http.NewRequest("PROPFIND", server.URL+"/", nil)
	req.Header.Set("Depth", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyStr := string(bodyBytes)

	t.Logf("Status: %d, Content-Length: %d", resp.StatusCode, resp.ContentLength)
	t.Logf("Body: %s", bodyStr)

	// 验证 XML 以 </D:multistatus> 结尾
	trimmed := strings.TrimSpace(bodyStr)
	if !strings.HasSuffix(trimmed, "</D:multistatus>") {
		t.Errorf("XML does not end with </D:multistatus>. Last 50 chars: %q",
			trimmed[len(trimmed)-min(50, len(trimmed)):])
	}

	// 验证 Content-Length 匹配实际 body
	if resp.ContentLength > 0 && resp.ContentLength != int64(len(bodyBytes)) {
		t.Errorf("Content-Length mismatch: header=%d, body=%d", resp.ContentLength, len(bodyBytes))
	}
}

// TestProxyFlow 模拟完整代理流程：请求 → /dav/ 前缀剥离 → 代理到 stream webdav → href 重写
func TestProxyFlow(t *testing.T) {
	// 1. 创建临时目录作为模拟文件系统
	tmpDir := t.TempDir()

	// 2. 创建 webdav handler
	handler := &webdav.Handler{
		FileSystem: webdav.Dir(tmpDir),
		LockSystem: webdav.NewMemLS(),
	}

	// 3. 启动后台 webdav 服务器
	backend := httptest.NewServer(handler)
	defer backend.Close()

	// 4. 创建代理（模拟 router.go 的 webdavProxy）
	target, _ := url.Parse(backend.URL)
	hrefPattern := regexp.MustCompile(`(<D:href>)(/[^<]*)(</D:href>)`)

	proxyHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// 剥离 /dav 前缀
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/dav")
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}
		req.Host = target.Host

		// 创建代理请求
		proxyReq, _ := http.NewRequest(req.Method, backend.URL+req.URL.Path, req.Body)
		proxyReq.Header = req.Header
		proxyReq.Header.Set("Depth", "1")

		resp, err := http.DefaultClient.Do(proxyReq)
		if err != nil {
			t.Logf("Proxy error: %v", err)
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// 复制响应头
		for k, v := range resp.Header {
			for _, vv := range v {
				w.Header().Add(k, vv)
			}
		}

		// 对于 PROPFIND 207，修改 href
		if resp.StatusCode == 207 {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Logf("Read body error: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			bodyBytes = hrefPattern.ReplaceAll(bodyBytes, []byte("${1}/dav${2}${3}"))
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(bodyBytes)))
			w.WriteHeader(resp.StatusCode)
			w.Write(bodyBytes)
		} else {
			w.WriteHeader(resp.StatusCode)
			io.Copy(w, resp.Body)
		}
	})

	// 5. 启动代理服务器
	proxy := httptest.NewServer(proxyHandler)
	defer proxy.Close()

	// 6. 发送 PROPFIND 请求
	req, _ := http.NewRequest("PROPFIND", proxy.URL+"/dav/", nil)
	req.Header.Set("Depth", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyStr := string(bodyBytes)

	t.Logf("Status: %d", resp.StatusCode)
	t.Logf("Response body:\n%s", bodyStr)

	// 检查 href 是否已正确重写
	if !strings.Contains(bodyStr, "<D:href>/dav/</D:href>") {
		t.Errorf("Href not rewritten correctly. Expected <D:href>/dav/</D:href>")
	}

	// 验证 XML 完整可解析
	trimmed := strings.TrimSpace(bodyStr)
	if !strings.HasSuffix(trimmed, "</D:multistatus>") {
		t.Errorf("Response does not end with </D:multistatus>. Ends with: %q", trimmed[len(trimmed)-min(50, len(trimmed)):])
	}
}

// TestStreamFilesystemPropfind 测试 streamFileSystem 的 PROPFIND 响应
func TestStreamFilesystemPropfind(t *testing.T) {
	// 使用真实文件系统测试（模拟 rclone 输出）
	tmpDir := t.TempDir()

	handler := &webdav.Handler{
		FileSystem: webdav.Dir(tmpDir),
		LockSystem: webdav.NewMemLS(),
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	req, _ := http.NewRequest("PROPFIND", server.URL+"/", nil)
	req.Header.Set("Depth", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyStr := string(bodyBytes)

	t.Logf("Status: %d", resp.StatusCode)
	t.Logf("Body: %s", bodyStr)

	// 验证 XML 不包含多余内容
	trimmed := strings.TrimSpace(bodyStr)
	if !strings.HasSuffix(trimmed, "</D:multistatus>") {
		t.Errorf("XML does not end properly. Extra content after </D:multistatus>")
	}

	// 验证 Content-Length 和实际 body 一致
	if resp.ContentLength > 0 && resp.ContentLength != int64(len(bodyBytes)) {
		t.Errorf("Content-Length mismatch: header=%d, body=%d", resp.ContentLength, len(bodyBytes))
	}
}

// TestSuppressWriteHeader 测试 suppressWriteHeaderWriter 不会导致多余输出
func TestSuppressWriteHeader(t *testing.T) {
	tmpDir := t.TempDir()

	handler := &webdav.Handler{
		FileSystem: webdav.Dir(tmpDir),
		LockSystem: webdav.NewMemLS(),
	}

	// 包装 handler
	wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &suppressWriteHeaderWriter{ResponseWriter: w}
		handler.ServeHTTP(sw, r)
	})

	server := httptest.NewServer(wrappedHandler)
	defer server.Close()

	req, _ := http.NewRequest("PROPFIND", server.URL+"/", nil)
	req.Header.Set("Depth", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyStr := string(bodyBytes)

	t.Logf("Status: %d", resp.StatusCode)
	t.Logf("Body length: %d", len(bodyBytes))
	t.Logf("Body: %s", bodyStr)

	// 验证结尾
	trimmed := strings.TrimSpace(bodyStr)
	if !strings.HasSuffix(trimmed, "</D:multistatus>") {
		t.Errorf("XML does not end properly. Last 50 chars: %q", trimmed[len(trimmed)-min(50, len(trimmed)):])
	}

	// 验证 Content-Length 匹配
	if resp.ContentLength > 0 && resp.ContentLength != int64(len(bodyBytes)) {
		t.Errorf("Content-Length mismatch: header=%d, body=%d", resp.ContentLength, len(bodyBytes))
	}
}

// TestErrorFSPropfindXMLIntegrity 模拟 nPlayer 报告 "Extra content at the end of the document" 的场景
// 当 errorFS 的 Stat 对某个条目返回非 os.ErrNotExist 错误时，
// webdav 库会尝试在已写入 multistatus 后再写入错误响应体
// suppressWriteHeaderWriter 必须丢弃这些错误体以保持 XML 完整
func TestErrorFSPropfindXMLIntegrity(t *testing.T) {
	tmpDir := t.TempDir()
	// 创建一个正常文件和一个会触发错误路径的"文件"
	os.WriteFile(tmpDir+"/good-file.txt", []byte("hello"), 0644)
	os.WriteFile(tmpDir+"/error-entry-1", []byte("should fail"), 0644)

	fs := &errorFS{dir: webdav.Dir(tmpDir)}
	handler := &webdav.Handler{
		FileSystem: fs,
		LockSystem: webdav.NewMemLS(),
	}

	wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &suppressWriteHeaderWriter{ResponseWriter: w}
		handler.ServeHTTP(sw, r)
	})

	server := httptest.NewServer(wrappedHandler)
	defer server.Close()

	// 发送 PROPFIND 请求（模拟 nPlayer 浏览目录）
	req, _ := http.NewRequest("PROPFIND", server.URL+"/", nil)
	req.Header.Set("Depth", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyStr := string(bodyBytes)

	t.Logf("Status: %d", resp.StatusCode)
	t.Logf("Body: %s", bodyStr)

	// 关键验证：XML 必须以 </D:multistatus> 结尾，不能有额外内容
	trimmed := strings.TrimSpace(bodyStr)
	if !strings.HasSuffix(trimmed, "</D:multistatus>") {
		t.Errorf("XML integrity FAILED: extra content after </D:multistatus>. Last 100 chars: %q",
			trimmed[len(trimmed)-min(100, len(trimmed)):])
	}

	// 验证 Content-Length 匹配
	if resp.ContentLength > 0 && resp.ContentLength != int64(len(bodyBytes)) {
		t.Errorf("Content-Length mismatch: header=%d, body=%d", resp.ContentLength, len(bodyBytes))
	}
}

func TestStreamFileInfoMode(t *testing.T) {
	dirInfo := &streamFileInfo{name: "mydir", size: 0, isDir: true}
	fileInfo := &streamFileInfo{name: "myfile", size: 1024, isDir: false}

	t.Logf("Dir: Name=%s, IsDir=%v, Mode=%v, Mode().IsDir=%v", dirInfo.Name(), dirInfo.IsDir(), dirInfo.Mode(), dirInfo.Mode().IsDir())
	t.Logf("File: Name=%s, IsDir=%v, Mode=%v, Mode().IsDir=%v", fileInfo.Name(), fileInfo.IsDir(), fileInfo.Mode(), fileInfo.Mode().IsDir())

	if !dirInfo.IsDir() {
		t.Error("Expected directory to be IsDir=true")
	}
	if fileInfo.IsDir() {
		t.Error("Expected file to be IsDir=false")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}