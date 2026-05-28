package adapter

import (
	"net/http"
	"time"
)

type RcloneConfig struct {
	BaseURL string
	User    string
	Pass    string
	Timeout time.Duration
}

type RcloneClient struct {
	config *RcloneConfig
	client *http.Client
}

type VersionResponse struct {
	Version    string `json:"version"`
	Decomposed []int  `json:"decomposed"`
	IsGit      bool   `json:"isGit"`
	IsBeta     bool   `json:"isBeta"`
	Os         string `json:"os"`
	OsKernel   string `json:"osKernel"`
	OsVersion  string `json:"osVersion"`
	OsArch     string `json:"osArch"`
	Arch       string `json:"arch"`
	GoVersion  string `json:"goVersion"`
	Linking    string `json:"linking"`
	GoTags     string `json:"goTags"`
}

type ListRemotesResponse struct {
	Remotes []string `json:"remotes"`
}

type CreateRemoteRequest struct {
	Name       string           `json:"name"`
	Type       string           `json:"type"`
	Parameters map[string]any   `json:"parameters"`
	Opt        *CreateRemoteOpt `json:"opt,omitempty"`
}

type CreateRemoteOpt struct {
	Obscure   bool `json:"obscure"`
	NoOutput  bool `json:"noOutput"`
	NoPrompts bool `json:"noPrompts"`
}

type DeleteRemoteRequest struct {
	Name string `json:"name"`
}

type UpdateRemoteRequest struct {
	Name       string         `json:"name"`
	Parameters map[string]any `json:"parameters"`
}

type DumpConfigResponse map[string]map[string]any

type GetConfigRequest struct {
	Name string `json:"name"`
}

type ProviderInfo struct {
	Name      string        `json:"Name"`
	Hangul    string        `json:"Hangul,omitempty"`
	Prefix    string        `json:"Prefix,omitempty"`
	OpenURL   string        `json:"OpenURL,omitempty"`
	HashTypes []string      `json:"HashTypes,omitempty"`
	Attr      []string      `json:"Attr,omitempty"`
	Policy    []string      `json:"Policy,omitempty"`
	Options   []OptionBlock `json:"Options,omitempty"`
}

type OptionBlock struct {
	Name       string    `json:"Name"`
	FieldName  string    `json:"FieldName,omitempty"`
	Help       string    `json:"Help"`
	Groups     string    `json:"Groups,omitempty"`
	Provider   string    `json:"Provider,omitempty"`
	Default    any       `json:"Default"`
	Value      any       `json:"Value,omitempty"`
	DefaultStr string    `json:"DefaultStr,omitempty"`
	ValueStr   string    `json:"ValueStr,omitempty"`
	Examples   []Example `json:"Examples,omitempty"`
	ShortOpt   string    `json:"ShortOpt,omitempty"`
	Hide       int       `json:"Hide"`
	Required   bool      `json:"Required"`
	IsPassword bool      `json:"IsPassword"`
	NoPrefix   bool      `json:"NoPrefix"`
	Advanced   bool      `json:"Advanced"`
	Exclusive  bool      `json:"Exclusive"`
	Sensitive  bool      `json:"Sensitive"`
}

type Example struct {
	Help  string `json:"Help"`
	Value any    `json:"Value"`
}

type ProvidersResponse struct {
	Providers []ProviderInfo `json:"providers"`
}

type PathInfo struct {
	Name     string         `json:"Name"`
	Path     string         `json:"Path"`
	IsDir    bool           `json:"IsDir"`
	IsBucket bool           `json:"IsBucket,omitempty"`
	MimeType string         `json:"MimeType,omitempty"`
	ModTime  string         `json:"ModTime,omitempty"`
	Size     int64          `json:"Size"`
	Hash     map[string]any `json:"Hash,omitempty"`
}

type ListPathRequest struct {
	Fs     string   `json:"fs"`
	Remote string   `json:"remote"`
	Opt    *ListOpt `json:"opt,omitempty"`
}

type ListOpt struct {
	Recurse      bool     `json:"recurse"`
	NoModTime    bool     `json:"noModTime"`
	DirsOnly     bool     `json:"dirsOnly"`
	FilesOnly    bool     `json:"filesOnly"`
	ShowHash     []string `json:"showHash,omitempty"`
	IgnoreSize   bool     `json:"ignoreSize"`
	IncludeEmpty bool     `json:"includeEmpty"`
	Flatten      []int    `json:"flatten,omitempty"`
}

type ListPathResponse struct {
	Fs   string     `json:"fs"`
	List []PathInfo `json:"list"`
}

type MkdirRequest struct {
	Fs     string `json:"fs"`
	Remote string `json:"remote"`
}

type DeleteFileRequest struct {
	Fs     string `json:"fs"`
	Remote string `json:"remote"`
}

type PurgeRequest struct {
	Fs     string `json:"fs"`
	Remote string `json:"remote"`
}

type MoveFileRequest struct {
	SrcFs     string `json:"srcFs"`
	SrcRemote string `json:"srcRemote"`
	DstFs     string `json:"dstFs"`
	DstRemote string `json:"dstRemote"`
}

type CopyFileRequest struct {
	SrcFs     string `json:"srcFs"`
	SrcRemote string `json:"srcRemote"`
	DstFs     string `json:"dstFs"`
	DstRemote string `json:"dstRemote"`
	Async     bool   `json:"_async"`
}

type AboutResponse struct {
	Used    int64 `json:"used"`
	Trashed int64 `json:"trashed,omitempty"`
	Other   int64 `json:"other,omitempty"`
	Free    int64 `json:"free"`
}

type FsRequest struct {
	Fs string `json:"fs"`
}

type FsInfoResponse struct {
	Name         string        `json:"Name"`
	Precision    int64         `json:"Precision"`
	Root         string        `json:"Root"`
	String       string        `json:"String"`
	Features     *Features     `json:"Features,omitempty"`
	Hashes       []string      `json:"Hashes,omitempty"`
	MetadataInfo *MetadataInfo `json:"MetadataInfo,omitempty"`
}

type Features struct {
	About                   bool `json:"About"`
	BucketBased             bool `json:"BucketBased"`
	BucketBasedRootOK       bool `json:"BucketBasedRootOK"`
	CanHaveEmptyDirectories bool `json:"CanHaveEmptyDirectories"`
	CaseInsensitive         bool `json:"CaseInsensitive"`
	ChangeNotify            bool `json:"ChangeNotify"`
	CleanUp                 bool `json:"CleanUp"`
	Command                 bool `json:"Command"`
	Copy                    bool `json:"Copy"`
	DirCacheFlush           bool `json:"DirCacheFlush"`
	DirMove                 bool `json:"DirMove"`
	Disconnect              bool `json:"Disconnect"`
	DuplicateFiles          bool `json:"DuplicateFiles"`
	GetTier                 bool `json:"GetTier"`
	IsLocal                 bool `json:"IsLocal"`
	ListR                   bool `json:"ListR"`
	MergeDirs               bool `json:"MergeDirs"`
	MetadataInfo            bool `json:"MetadataInfo"`
	Move                    bool `json:"Move"`
	OpenWriterAt            bool `json:"OpenWriterAt"`
	PublicLink              bool `json:"PublicLink"`
	Purge                   bool `json:"Purge"`
	PutStream               bool `json:"PutStream"`
	PutUnchecked            bool `json:"PutUnchecked"`
	ReadMetadata            bool `json:"ReadMetadata"`
	ReadMimeType            bool `json:"ReadMimeType"`
	ServerSideAcrossConfigs bool `json:"ServerSideAcrossConfigs"`
	SetTier                 bool `json:"SetTier"`
	SetWrapper              bool `json:"SetWrapper"`
	Shutdown                bool `json:"Shutdown"`
	SlowHash                bool `json:"SlowHash"`
	SlowModTime             bool `json:"SlowModTime"`
	UnWrap                  bool `json:"UnWrap"`
	UserInfo                bool `json:"UserInfo"`
	UserMetadata            bool `json:"UserMetadata"`
	WrapFs                  bool `json:"WrapFs"`
	WriteMetadata           bool `json:"WriteMetadata"`
	WriteMimeType           bool `json:"WriteMimeType"`
}

type MetadataInfo struct {
	System map[string]MetadataField `json:"System,omitempty"`
}

type MetadataField struct {
	Help    string `json:"Help"`
	Type    string `json:"Type"`
	Example string `json:"Example,omitempty"`
}

type PublicLinkResponse struct {
	URL string `json:"url"`
}

type PublicLinkRequest struct {
	Fs          string `json:"fs"`
	Remote      string `json:"remote"`
	Expire      string `json:"expire,omitempty"`
	Unpublished bool   `json:"unpublished"`
}

type SyncCopyRequest struct {
	SrcFs              string `json:"srcFs"`
	DstFs              string `json:"dstFs"`
	CreateEmptySrcDirs bool   `json:"createEmptySrcDirs"`
}

type SyncMoveRequest struct {
	SrcFs              string `json:"srcFs"`
	DstFs              string `json:"dstFs"`
	CreateEmptySrcDirs bool   `json:"createEmptySrcDirs"`
	DeleteEmptySrcDirs bool   `json:"deleteEmptySrcDirs"`
}

type JobIDResponse struct {
	JobID int64 `json:"jobid"`
}

type StartJobRequest struct {
	SrcFs string `json:"srcFs"`
	DstFs string `json:"dstFs"`
	Async bool   `json:"_async"`
	*TaskOptions
}

type jobStatusResponse struct {
	ID        int64          `json:"id"`
	ExecuteID string         `json:"executeId"`
	StartTime string         `json:"startTime"`
	EndTime   string         `json:"endTime,omitempty"`
	Duration  float64        `json:"duration"`
	Success   bool           `json:"success"`
	Finished  bool           `json:"finished"`
	Error     string         `json:"error,omitempty"`
	Output    map[string]any `json:"output,omitempty"`
	Progress  map[string]any `json:"progress,omitempty"`
}

type JobIDRequest struct {
	JobID int64 `json:"jobid"`
}