package app

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "path"
    "strconv"
    "strings"

    "rcloneflow/internal/rclone"
    "rcloneflow/internal/scheduler"
    "rcloneflow/internal/store"
)

type Server struct {
    rc *rclone.Client
    db *store.DB
}

func Run() error {
    dbDir := os.Getenv("APP_DATA_DIR")
    if dbDir == "" { dbDir = "./data" }
    db, err := store.Open(dbDir)
    if err != nil { return err }
    srv := &Server{rc: rclone.NewFromEnv(), db: db}
    sched := scheduler.New(db, srv)
    if err := sched.Start(); err != nil { return err }

    mux := http.NewServeMux()
    mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { json.NewEncoder(w).Encode(map[string]any{"ok": true}) })
    mux.HandleFunc("/api/remotes", srv.handleRemotes)
    mux.HandleFunc("/api/remotes/config/", srv.handleRemoteConfig)
    mux.HandleFunc("/api/remotes/test", srv.handleRemoteTest)
    mux.HandleFunc("/api/browser/list", srv.handleBrowser)
    mux.HandleFunc("/api/providers", srv.handleProviders)
    mux.HandleFunc("/api/config/dump", srv.handleConfigDump)
    mux.HandleFunc("/api/config/", srv.handleConfigActions)
    mux.HandleFunc("/api/usage/", srv.handleUsage)
    mux.HandleFunc("/api/fsinfo/", srv.handleFsInfo)
    mux.HandleFunc("/api/fs/mkdir", srv.handleMkdir)
    mux.HandleFunc("/api/fs/delete", srv.handleDeleteFile)
    mux.HandleFunc("/api/fs/purge", srv.handlePurge)
    mux.HandleFunc("/api/fs/move", srv.handleMove)
    mux.HandleFunc("/api/fs/copy", srv.handleCopy)
    mux.HandleFunc("/api/fs/publiclink", srv.handlePublicLink)
    mux.HandleFunc("/api/tasks", srv.handleTasks)
    mux.HandleFunc("/api/tasks/", srv.handleTaskActions)
    mux.HandleFunc("/api/schedules", srv.handleSchedules)
    mux.HandleFunc("/api/runs", srv.handleRuns)
    mux.HandleFunc("/api/runs/", srv.handleRunStatus)

    mux.Handle("/", http.FileServer(http.Dir("./web")))

    addr := os.Getenv("APP_ADDR")
    if addr == "" { addr = ":17870" }
    log.Printf("listening on %s", addr)
    return http.ListenAndServe(addr, withCORS(mux))
}

func withCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
        if r.Method == http.MethodOptions { return }
        next.ServeHTTP(w, r)
    })
}

func writeJSON(w http.ResponseWriter, code int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    _ = json.NewEncoder(w).Encode(v)
}

func decode(r *http.Request, dst any) error { return json.NewDecoder(r.Body).Decode(dst) }

func (s *Server) handleRemotes(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        remotes, err := s.rc.ListRemotes(r.Context())
        if err != nil { writeJSON(w, 500, map[string]any{"error": err.Error()}); return }
        version, _ := s.rc.Version(r.Context())
        writeJSON(w, 200, map[string]any{"remotes": remotes, "version": version})
    case http.MethodPost:
        var req struct {
            Name       string            `json:"name"`
            Type       string            `json:"type"`
            Parameters map[string]any    `json:"parameters"`
        }
        if err := decode(r, &req); err != nil { writeJSON(w, 400, map[string]any{"error": err.Error()}); return }
        if err := s.rc.CreateRemote(r.Context(), req.Name, req.Type, req.Parameters); err != nil {
            writeJSON(w, 500, map[string]any{"error": err.Error()}); return
        }
        writeJSON(w, 200, map[string]any{"created": true})
    case http.MethodPut:
        // 更新存储配置
        var req struct {
            Name       string            `json:"name"`
            Type       string            `json:"type"`
            Parameters map[string]any    `json:"parameters"`
        }
        if err := decode(r, &req); err != nil { writeJSON(w, 400, map[string]any{"error": err.Error()}); return }
        if err := s.rc.CreateRemote(r.Context(), req.Name, req.Type, req.Parameters); err != nil {
            writeJSON(w, 500, map[string]any{"error": err.Error()}); return
        }
        writeJSON(w, 200, map[string]any{"updated": true})
    default:
        w.WriteHeader(405)
    }
}

// handleRemoteConfig 获取单个存储配置
func (s *Server) handleRemoteConfig(w http.ResponseWriter, r *http.Request) {
    name := strings.TrimPrefix(r.URL.Path, "/api/remotes/config/")
    if name == "" {
        writeJSON(w, 400, map[string]any{"error": "name required"})
        return
    }
    if r.Method != http.MethodGet {
        w.WriteHeader(405)
        return
    }
    cfg, err := s.rc.GetConfig(r.Context(), name)
    if err != nil {
        writeJSON(w, 500, map[string]any{"error": err.Error()})
        return
    }
    writeJSON(w, 200, cfg)
}

func (s *Server) handleRemoteTest(w http.ResponseWriter, r *http.Request) {
    var req struct{ Name string `json:"name"` }
    if err := decode(r, &req); err != nil { writeJSON(w, 400, map[string]any{"error": err.Error()}); return }
    items, err := s.rc.ListPath(r.Context(), req.Name+":", "")
    if err != nil { writeJSON(w, 500, map[string]any{"error": err.Error()}); return }
    writeJSON(w, 200, map[string]any{"ok": true, "count": len(items)})
}

func (s *Server) handleBrowser(w http.ResponseWriter, r *http.Request) {
    remote := r.URL.Query().Get("remote")
    p := strings.Trim(strings.TrimPrefix(r.URL.Query().Get("path"), "/"), " ")
    if p == "." { p = "" }
    fsPath := remote + ":"
    items, err := s.rc.ListPath(r.Context(), fsPath, p)
    if err != nil { writeJSON(w, 500, map[string]any{"error": err.Error()}); return }
    current := fsPath
    if p != "" { current += path.Clean("/" + p) }
    writeJSON(w, 200, map[string]any{"fs": current, "items": items})
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        writeJSON(w, 200, s.db.ListTasks())
    case http.MethodPost:
        var req store.Task
        if err := decode(r, &req); err != nil { writeJSON(w, 400, map[string]any{"error": err.Error()}); return }
        t, err := s.db.AddTask(req)
        if err != nil { writeJSON(w, 500, map[string]any{"error": err.Error()}); return }
        writeJSON(w, 200, t)
    default:
        w.WriteHeader(405)
    }
}

func (s *Server) RunTask(ctx context.Context, taskID int64, trigger string) error {
    t, ok := s.db.GetTask(taskID)
    if !ok { return fmt.Errorf("task not found") }
    src := t.SourceRemote + ":" + strings.TrimPrefix(t.SourcePath, "/")
    dst := t.TargetRemote + ":" + strings.TrimPrefix(t.TargetPath, "/")
    jobID, err := s.rc.StartJob(ctx, t.Mode, src, dst)
    run, _ := s.db.AddRun(store.Run{TaskID: taskID, RcJobID: jobID, Status: "running", Trigger: trigger})
    if err != nil {
        _ = s.db.UpdateRun(run.ID, func(r *store.Run) { r.Status = "failed"; r.Error = err.Error() })
        return err
    }
    return nil
}

func (s *Server) handleTaskActions(w http.ResponseWriter, r *http.Request) {
    p := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
    if !strings.HasSuffix(p, "/run") { w.WriteHeader(404); return }
    idStr := strings.TrimSuffix(p, "/run")
    id, _ := strconv.ParseInt(strings.Trim(idStr, "/"), 10, 64)
    if err := s.RunTask(r.Context(), id, "manual"); err != nil {
        writeJSON(w, 500, map[string]any{"error": err.Error()}); return
    }
    writeJSON(w, 200, map[string]any{"started": true})
}

func (s *Server) handleSchedules(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        writeJSON(w, 200, s.db.ListSchedules())
    case http.MethodPost:
        var req store.Schedule
        if err := decode(r, &req); err != nil { writeJSON(w, 400, map[string]any{"error": err.Error()}); return }
        item, err := s.db.AddSchedule(req)
        if err != nil { writeJSON(w, 500, map[string]any{"error": err.Error()}); return }
        writeJSON(w, 200, item)
    default:
        w.WriteHeader(405)
    }
}

func (s *Server) handleRuns(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, 200, s.db.ListRuns())
}

func (s *Server) handleRunStatus(w http.ResponseWriter, r *http.Request) {
    id, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/runs/"), 10, 64)
    for _, run := range s.db.ListRuns() {
        if run.ID != id { continue }
        if run.RcJobID > 0 {
            st, err := s.rc.JobStatus(r.Context(), run.RcJobID)
            if err == nil {
                _ = s.db.UpdateRun(run.ID, func(rr *store.Run) {
                    rr.Summary = st
                    if finished, ok := st["finished"].(bool); ok && finished {
                        rr.Status = "finished"
                    }
                    if success, ok := st["success"].(bool); ok && !success {
                        rr.Status = "failed"
                    }
                })
                writeJSON(w, 200, st)
                return
            }
        }
        writeJSON(w, 200, run)
        return
    }
    writeJSON(w, 404, map[string]any{"error": "run not found"})
}

// handleProviders 获取所有存储提供商
func (s *Server) handleProviders(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(405)
        return
    }
    providers, err := s.rc.GetProviders(r.Context())
    if err != nil {
        writeJSON(w, 500, map[string]any{"error": err.Error()})
        return
    }
    writeJSON(w, 200, map[string]any{"providers": providers})
}

// handleConfigDump 获取所有存储配置
func (s *Server) handleConfigDump(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(405)
        return
    }
    config, err := s.rc.DumpConfig(r.Context())
    if err != nil {
        writeJSON(w, 500, map[string]any{"error": err.Error()})
        return
    }
    writeJSON(w, 200, config)
}

// handleConfigActions 获取/删除单个存储配置
func (s *Server) handleConfigActions(w http.ResponseWriter, r *http.Request) {
    name := strings.TrimPrefix(r.URL.Path, "/api/config/")
    
    switch r.Method {
    case http.MethodGet:
        cfg, err := s.rc.GetConfig(r.Context(), name)
        if err != nil {
            writeJSON(w, 500, map[string]any{"error": err.Error()})
            return
        }
        writeJSON(w, 200, cfg)
    case http.MethodDelete:
        if err := s.rc.DeleteRemote(r.Context(), name); err != nil {
            writeJSON(w, 500, map[string]any{"error": err.Error()})
            return
        }
        writeJSON(w, 200, map[string]any{"deleted": true})
    default:
        w.WriteHeader(405)
    }
}

// handleUsage 获取存储使用量
func (s *Server) handleUsage(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(405)
        return
    }
    fs := strings.TrimPrefix(r.URL.Path, "/api/usage/")
    if fs == "" {
        writeJSON(w, 400, map[string]any{"error": "fs parameter required"})
        return
    }
    usage, err := s.rc.GetUsage(r.Context(), fs)
    if err != nil {
        writeJSON(w, 500, map[string]any{"error": err.Error()})
        return
    }
    writeJSON(w, 200, usage)
}

// handleFsInfo 获取文件系统信息
func (s *Server) handleFsInfo(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(405)
        return
    }
    fs := strings.TrimPrefix(r.URL.Path, "/api/fsinfo/")
    if fs == "" {
        writeJSON(w, 400, map[string]any{"error": "fs parameter required"})
        return
    }
    info, err := s.rc.GetFsInfo(r.Context(), fs)
    if err != nil {
        writeJSON(w, 500, map[string]any{"error": err.Error()})
        return
    }
    writeJSON(w, 200, info)
}

// handleMkdir 创建目录
func (s *Server) handleMkdir(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(405)
        return
    }
    var req struct {
        Fs     string `json:"fs"`
        Remote string `json:"remote"`
    }
    if err := decode(r, &req); err != nil {
        writeJSON(w, 400, map[string]any{"error": err.Error()})
        return
    }
    if err := s.rc.Mkdir(r.Context(), req.Fs, req.Remote); err != nil {
        writeJSON(w, 500, map[string]any{"error": err.Error()})
        return
    }
    writeJSON(w, 200, map[string]any{"ok": true})
}

// handleDeleteFile 删除文件
func (s *Server) handleDeleteFile(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(405)
        return
    }
    var req struct {
        SrcFs     string `json:"srcFs"`
        SrcRemote string `json:"srcRemote"`
    }
    if err := decode(r, &req); err != nil {
        writeJSON(w, 400, map[string]any{"error": err.Error()})
        return
    }
    if err := s.rc.DeleteFile(r.Context(), req.SrcFs, req.SrcRemote); err != nil {
        writeJSON(w, 500, map[string]any{"error": err.Error()})
        return
    }
    writeJSON(w, 200, map[string]any{"ok": true})
}

// handlePurge 删除目录
func (s *Server) handlePurge(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(405)
        return
    }
    var req struct {
        Fs     string `json:"fs"`
        Remote string `json:"remote"`
    }
    if err := decode(r, &req); err != nil {
        writeJSON(w, 400, map[string]any{"error": err.Error()})
        return
    }
    if err := s.rc.Purge(r.Context(), req.Fs, req.Remote); err != nil {
        writeJSON(w, 500, map[string]any{"error": err.Error()})
        return
    }
    writeJSON(w, 200, map[string]any{"ok": true})
}

// handleMove 移动/重命名
func (s *Server) handleMove(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(405)
        return
    }
    var req struct {
        SrcFs     string `json:"srcFs"`
        SrcRemote string `json:"srcRemote"`
        DstFs     string `json:"dstFs"`
        DstRemote string `json:"dstRemote"`
    }
    if err := decode(r, &req); err != nil {
        writeJSON(w, 400, map[string]any{"error": err.Error()})
        return
    }
    if err := s.rc.MoveFile(r.Context(), req.SrcFs, req.SrcRemote, req.DstFs, req.DstRemote); err != nil {
        writeJSON(w, 500, map[string]any{"error": err.Error()})
        return
    }
    writeJSON(w, 200, map[string]any{"ok": true})
}

// handleCopy 复制
func (s *Server) handleCopy(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(405)
        return
    }
    var req struct {
        SrcFs     string `json:"srcFs"`
        SrcRemote string `json:"srcRemote"`
        DstFs     string `json:"dstFs"`
        DstRemote string `json:"dstRemote"`
    }
    if err := decode(r, &req); err != nil {
        writeJSON(w, 400, map[string]any{"error": err.Error()})
        return
    }
    if err := s.rc.CopyFile(r.Context(), req.SrcFs, req.SrcRemote, req.DstFs, req.DstRemote); err != nil {
        writeJSON(w, 500, map[string]any{"error": err.Error()})
        return
    }
    writeJSON(w, 200, map[string]any{"ok": true})
}

// handlePublicLink 生成分享链接
func (s *Server) handlePublicLink(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(405)
        return
    }
    var req struct {
        Fs     string `json:"fs"`
        Remote string `json:"remote"`
    }
    if err := decode(r, &req); err != nil {
        writeJSON(w, 400, map[string]any{"error": err.Error()})
        return
    }
    url, err := s.rc.PublicLink(r.Context(), req.Fs, req.Remote)
    if err != nil {
        writeJSON(w, 500, map[string]any{"error": err.Error()})
        return
    }
    writeJSON(w, 200, map[string]any{"url": url})
}
