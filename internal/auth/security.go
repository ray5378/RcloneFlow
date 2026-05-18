package auth

import (
	"net/http"
	"regexp"
	"strings"
)

// PathSecurityMiddleware 路径安全中间件 - 检查路径穿越攻击
func PathSecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 获取路径参数
		path := r.URL.Query().Get("path")
		if path == "" {
			// 从请求体获取
			next.ServeHTTP(w, r)
			return
		}

		// 检查路径穿越
		if containsPathTraversal(path) {
			http.Error(w, `{"error":"路径包含非法字符"}`, http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// containsPathTraversal 检查是否存在路径穿越
func containsPathTraversal(path string) bool {
	// 规范化路径（统一使用正斜杠）
	normalized := strings.ReplaceAll(path, "\\", "/")

	// 检查 ..
	if strings.Contains(normalized, "..") {
		return true
	}

	// 检查 ./ 或 开头/结尾的.
	if strings.HasSuffix(normalized, "/.") || strings.HasSuffix(normalized, ".") {
		// 允许单.作为当前目录
		if normalized == "." || normalized == "./" {
			return false
		}
		// 检查类似 .hidden 或 .txt 等合法文件名（不以/开头）
		if strings.HasPrefix(normalized, ".") && len(normalized) > 1 {
			// 检查第二字符是否是分隔符
			if len(normalized) > 2 && normalized[1] == '/' {
				return true // .隐藏文件/xxx 这种不应该出现
			}
		}
	}

	// 检查 %2e 编码的 ..
	lowerPath := strings.ToLower(path)
	if strings.Contains(lowerPath, "%2e%2e") || strings.Contains(lowerPath, "%2e.") {
		return true
	}

	// 检查空字节攻击
	if strings.Contains(path, "\x00") {
		return true
	}

	return false
}

// ValidatePath 验证路径安全性
func ValidatePath(path string) bool {
	return !containsPathTraversal(path)
}

// SanitizePath 清理路径中的潜在危险字符
func SanitizePath(path string) string {
	// 移除 null 字节
	path = strings.ReplaceAll(path, "\x00", "")
	// 移除 \.. 模式
	re := regexp.MustCompile(`\\+\.\.`)
	path = re.ReplaceAllString(path, "")
	// 移除 /.. 模式
	re = regexp.MustCompile(`/+\.\.`)
	path = re.ReplaceAllString(path, "")
	return path
}
