package rclone

// StaticProviders 返回 provider 元数据（最小子集），用于前端驱动「添加存储」表单。
// 说明：这里先给出常见 provider 的关键字段（必填/可选/高级），后续可随 rclone 版本扩充/生成。
func StaticProviders() []map[string]any {
	return []map[string]any{
		{
			"Name":        "webdav",
			"Description": "WebDAV-compatible storage",
			"Options": []map[string]any{
				{"Name": "url", "Help": "WebDAV endpoint URL", "Required": true},
				{"Name": "vendor", "Help": "WebDAV vendor", "Required": false, "Examples": []map[string]any{{"Value": "other"}}},
				{"Name": "user", "Help": "Username", "Required": false},
				{"Name": "pass", "Help": "Password", "IsPassword": true, "Required": false},
			},
		},
		{
			"Name":        "smb",
			"Description": "SMB/CIFS network file system",
			"Options": []map[string]any{
				{"Name": "host", "Help": "SMB host", "Required": true},
				{"Name": "port", "Help": "Port", "Required": false, "DefaultStr": "445"},
				{"Name": "user", "Help": "Username", "Required": false},
				{"Name": "pass", "Help": "Password", "IsPassword": true, "Required": false},
			},
		},
		{
			"Name":        "s3",
			"Description": "Amazon S3 and compatible",
			"Options": []map[string]any{
				{"Name": "provider", "Help": "S3 provider", "Required": true, "Examples": []map[string]any{{"Value": "AWS"}, {"Value": "Other"}}},
				{"Name": "access_key_id", "Help": "Access key", "Required": false},
				{"Name": "secret_access_key", "Help": "Secret key", "IsPassword": true, "Required": false},
				{"Name": "endpoint", "Help": "Custom endpoint", "Required": false},
			},
		},
	}
}
