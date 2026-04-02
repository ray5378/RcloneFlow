package adapter

// TaskOptions rclone任务选项
type TaskOptions struct {
	// 过滤参数 (Filtering)
	Exclude            []string `json:"exclude,omitempty"`            // --exclude
	ExcludeFrom        []string `json:"excludeFrom,omitempty"`        // --exclude-from
	ExcludeIfPresent   []string `json:"excludeIfPresent,omitempty"`   // --exclude-if-present
	Include            []string `json:"include,omitempty"`            // --include
	IncludeFrom        []string `json:"includeFrom,omitempty"`        // --include-from
	Filter             []string `json:"filter,omitempty"`             // --filter
	FilterFrom         []string `json:"filterFrom,omitempty"`         // --filter-from
	FilesFrom          []string `json:"filesFrom,omitempty"`          // --files-from
	FilesFromRaw       []string `json:"filesFromRaw,omitempty"`       // --files-from-raw
	DeleteExcluded     bool     `json:"deleteExcluded,omitempty"`     // --delete-excluded
	IgnoreCase         bool     `json:"ignoreCase,omitempty"`         // --ignore-case
	IgnoreCaseSync     bool     `json:"ignoreCaseSync,omitempty"`     // --ignore-case-sync
	IgnoreExisting     bool     `json:"ignoreExisting,omitempty"`     // --ignore-existing

	// 比较策略 (Compare)
	Checksum         bool   `json:"checksum,omitempty"`         // --checksum
	SizeOnly         bool   `json:"sizeOnly,omitempty"`         // --size-only
	IgnoreSize       bool   `json:"ignoreSize,omitempty"`       // --ignore-size
	IgnoreTimes      bool   `json:"ignoreTimes,omitempty"`      // --ignore-times
	Update           bool   `json:"update,omitempty"`           // --update
	UseServerModtime bool   `json:"useServerModtime,omitempty"` // --use-server-modtime
	ModifyWindow     string `json:"modifyWindow,omitempty"`     // --modify-window
	RefreshTimes     bool   `json:"refreshTimes,omitempty"`     // --refresh-times

	// 路径策略 (Path)
	NoTraverse   bool   `json:"noTraverse,omitempty"`   // --no-traverse
	NoCheckDest  bool   `json:"noCheckDest,omitempty"`  // --no-check-dest
	CompareDest  string `json:"compareDest,omitempty"`  // --compare-dest
	CopyDest     string `json:"copyDest,omitempty"`     // --copy-dest

	// 元数据 (Metadata)
	IgnoreChecksum bool `json:"ignoreChecksum,omitempty"` // --ignore-checksum

	// 多线程 (Multi-threading)
	MultiThreadStreams       int `json:"multiThreadStreams,omitempty"`        // --multi-thread-streams
	MultiThreadCutoff        int `json:"multiThreadCutoff,omitempty"`         // --multi-thread-cutoff
	MultiThreadWriteBufferSize int `json:"multiThreadWriteBufferSize,omitempty"` // --multi-thread-write-buffer-size

	// 其他复制标志
	Inplace     bool   `json:"inplace,omitempty"`     // --inplace
	Immutable   bool   `json:"immutable,omitempty"`   // --immutable
	OrderBy     string `json:"orderBy,omitempty"`     // --order-by
	MaxTransfer int64  `json:"maxTransfer,omitempty"` // --max-transfer
	MaxDuration int    `json:"maxDuration,omitempty"` // --max-duration (seconds)
	CheckFirst  bool   `json:"checkFirst,omitempty"`  // --check-first

	// 备份 (Backup)
	BackupDir string `json:"backupDir,omitempty"` // --backup-dir
	Suffix    string `json:"suffix,omitempty"`    // --suffix

	// 传输控制
	CutoffMode string `json:"cutoffMode,omitempty"` // --cutoff-mode

	// 同步标志 (Sync)
	DeleteBefore  bool `json:"deleteBefore,omitempty"`  // --delete-before
	DeleteDuring  bool `json:"deleteDuring,omitempty"`  // --delete-during
	DeleteAfter   bool `json:"deleteAfter,omitempty"`   // --delete-after
	MaxDelete     int  `json:"maxDelete,omitempty"`     // --max-delete
	MaxDeleteSize int64 `json:"maxDeleteSize,omitempty"` // --max-delete-size
	TrackRenames  bool `json:"trackRenames,omitempty"`  // --track-renames
	IgnoreErrors bool `json:"ignoreErrors,omitempty"`   // --ignore-errors

	// 网络参数 (Networking)
	BwLimit            string `json:"bwLimit,omitempty"`             // --bwlimit
	BwLimitFile        string `json:"bwLimitFile,omitempty"`         // --bwlimit-file
	Bind               string `json:"bind,omitempty"`               // --bind
	ConnTimeout        int    `json:"connTimeout,omitempty"`         // --contimeout (seconds)
	Timeout            int    `json:"timeout,omitempty"`             // --timeout (seconds)
	ExpectContinueTimeout int `json:"expectContinueTimeout,omitempty"` // --expect-continue-timeout (seconds)
	Header             map[string]string `json:"header,omitempty"`    // --header
	HeaderDownload     map[string]string `json:"headerDownload,omitempty"` // --header-download
	HeaderUpload       map[string]string `json:"headerUpload,omitempty"`   // --header-upload
	CaCert             string `json:"caCert,omitempty"`             // --ca-cert
	ClientCert         string `json:"clientCert,omitempty"`         // --client-cert
	ClientKey          string `json:"clientKey,omitempty"`          // --client-key
	NoCheckCertificate bool   `json:"noCheckCertificate,omitempty"` // --no-check-certificate
	TpsLimit           float64 `json:"tpsLimit,omitempty"`         // --tpslimit
	TpsLimitBurst      int     `json:"tpsLimitBurst,omitempty"`     // --tpslimit-burst
	UserAgent          string `json:"userAgent,omitempty"`         // --user-agent
	UseCookies         bool   `json:"useCookies,omitempty"`        // --use-cookies
	DisableHttpKeepAlives bool `json:"disableHttpKeepAlives,omitempty"` // --disable-http-keep-alives
	Dscp               string `json:"dscp,omitempty"`              // --dscp

	// 性能参数 (Performance)
	Transfers   int `json:"transfers,omitempty"`   // --transfers
	Checkers    int `json:"checkers,omitempty"`    // --checkers
	BufferSize  int `json:"bufferSize,omitempty"`  // --buffer-size (MB)

	// 日志输出 (Logging)
	Verbose        bool   `json:"verbose,omitempty"`         // --verbose
	Quiet          bool   `json:"quiet,omitempty"`           // --quiet
	LogFile        string `json:"logFile,omitempty"`         // --log-file
	LogFormat      string `json:"logFormat,omitempty"`       // --log-format
	HumanReadable  bool   `json:"humanReadable,omitempty"`   // --human-readable
	UseJsonLog     bool   `json:"useJsonLog,omitempty"`      // --use-json-log

	// 配置参数 (Config)
	ConfigDir     string `json:"configDir,omitempty"`     // --config
	CacheDir      string `json:"cacheDir,omitempty"`      // --cache-dir
	TempDir       string `json:"tempDir,omitempty"`       // --temp-dir
	Interactive   bool   `json:"interactive,omitempty"`   // --interactive
	DryRun        bool   `json:"dryRun,omitempty"`        // --dry-run
	AutoConfirm   bool   `json:"autoConfirm,omitempty"`   // --auto-confirm
	ErrorOnNoTransfer bool `json:"errorOnNoTransfer,omitempty"` // --error-on-no-transfer
	Retries       int    `json:"retries,omitempty"`       // --retries
	LowLevelRetries int   `json:"lowLevelRetries,omitempty"` // --low-level-retries
	AskPassword   bool   `json:"askPassword,omitempty"`   // --ask-password
	PasswordCommand string `json:"passwordCommand,omitempty"` // --password-command
	UseMmap       bool   `json:"useMmap,omitempty"`       // --use-mmap
	NoUnicodeNormalization bool `json:"noUnicodeNormalization,omitempty"` // --no-unicode-normalization
	Color         string `json:"color,omitempty"`         // --color

	// 其他
	ServerSideAcrossConfigs bool `json:"serverSideAcrossConfigs,omitempty"` // --server-side-across-configs
}

// IsEmpty 检查是否没有设置任何选项
func (o *TaskOptions) IsEmpty() bool {
	if o == nil {
		return true
	}
	// 检查关键字段是否都是零值
	return o.Exclude == nil &&
		o.ExcludeFrom == nil &&
		o.ExcludeIfPresent == nil &&
		o.Include == nil &&
		o.IncludeFrom == nil &&
		o.Filter == nil &&
		o.FilterFrom == nil &&
		o.FilesFrom == nil &&
		o.FilesFromRaw == nil &&
		!o.DeleteExcluded &&
		!o.IgnoreCase &&
		!o.IgnoreCaseSync &&
		!o.IgnoreExisting &&
		!o.Checksum &&
		!o.SizeOnly &&
		!o.IgnoreSize &&
		!o.IgnoreTimes &&
		!o.Update &&
		!o.UseServerModtime &&
		o.ModifyWindow == "" &&
		!o.RefreshTimes &&
		!o.NoTraverse &&
		!o.NoCheckDest &&
		o.CompareDest == "" &&
		o.CopyDest == "" &&
		!o.IgnoreChecksum &&
		o.MultiThreadStreams == 0 &&
		o.MultiThreadCutoff == 0 &&
		o.MultiThreadWriteBufferSize == 0 &&
		!o.Inplace &&
		!o.Immutable &&
		o.OrderBy == "" &&
		o.MaxTransfer == 0 &&
		o.MaxDuration == 0 &&
		!o.CheckFirst &&
		o.BackupDir == "" &&
		o.Suffix == "" &&
		o.CutoffMode == "" &&
		!o.DeleteBefore &&
		!o.DeleteDuring &&
		!o.DeleteAfter &&
		o.MaxDelete == 0 &&
		o.MaxDeleteSize == 0 &&
		!o.TrackRenames &&
		!o.IgnoreErrors &&
		o.BwLimit == "" &&
		o.BwLimitFile == "" &&
		o.Bind == "" &&
		o.ConnTimeout == 0 &&
		o.Timeout == 0 &&
		o.ExpectContinueTimeout == 0 &&
		o.Header == nil &&
		o.HeaderDownload == nil &&
		o.HeaderUpload == nil &&
		o.CaCert == "" &&
		o.ClientCert == "" &&
		o.ClientKey == "" &&
		!o.NoCheckCertificate &&
		o.TpsLimit == 0 &&
		o.TpsLimitBurst == 0 &&
		o.UserAgent == "" &&
		!o.UseCookies &&
		!o.DisableHttpKeepAlives &&
		o.Dscp == "" &&
		o.Transfers == 0 &&
		o.Checkers == 0 &&
		o.BufferSize == 0 &&
		!o.Verbose &&
		!o.Quiet &&
		o.LogFile == "" &&
		o.LogFormat == "" &&
		!o.HumanReadable &&
		!o.UseJsonLog &&
		o.ConfigDir == "" &&
		o.CacheDir == "" &&
		o.TempDir == "" &&
		!o.Interactive &&
		!o.DryRun &&
		!o.AutoConfirm &&
		!o.ErrorOnNoTransfer &&
		o.Retries == 0 &&
		o.LowLevelRetries == 0 &&
		!o.AskPassword &&
		o.PasswordCommand == "" &&
		!o.UseMmap &&
		!o.NoUnicodeNormalization &&
		o.Color == "" &&
		!o.ServerSideAcrossConfigs
}
