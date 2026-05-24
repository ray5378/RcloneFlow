# Bisync 双向同步集成设计文档

## 一、概述

将 rclone 的 `bisync` 命令集成到现有的任务系统中，支持双向同步、lst 文件管理、配置项管理等功能。

## 二、功能需求

1. **任务模式新增**: 新增 `bisync` 任务模式，仅在非 CAS 兼容链路可用
2. **Bisync 配置弹窗**:
   - Resync 按钮
   - 基本配置项
   - 高级配置项（折叠）
   - LST 文件列表：
     - 显示历史版本
     - 删除功能
     - 回滚功能（点击某个版本即可回滚）
3. **运行时显示**: 传输中弹窗与历史详情需区分显示 A→B 和 B→A 两个方向
4. **数据清理**: 删除任务或模式从 bisync 切换到其他时，删除对应 bisync 目录
5. **各模块补充**: 标签、历史、导入导出、日志清理等模块均需支持 bisync

## 三、系统架构

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  前端层                                                                    │
│  ┌────────────────────┐  ┌─────────────────────────┐  ┌─────────────────┐  │
│  │ BisyncConfigModal │  │ AddTaskForm (bisync)    │  │ TaskView (UI)   │  │
│  └──────────┬─────────┘  └────────────┬────────────┘  └────────┬────────┘  │
│             │                         │                         │            │
│  ┌──────────▼─────────────────────────▼─────────────────────────▼────────┐  │
│  │  frontend/src/api/task.ts (新增 bisync 接口)                          │  │
│  └──────────┬────────────────────────────────────────────────────────────┘  │
└─────────────┼───────────────────────────────────────────────────────────────┘
              │
┌─────────────▼───────────────────────────────────────────────────────────────┐
│  后端控制层                                                                │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │ internal/controller/task.go (新增 bisync 端点)                        │ │
│  └──────────┬────────────────────────────────────────────────────────────┘ │
└─────────────┼───────────────────────────────────────────────────────────────┘
              │
┌─────────────▼───────────────────────────────────────────────────────────────┐
│  后端服务层                                                                │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │ internal/service/task.go            (bisync 目录管理)                 │ │
│  │ internal/service/tag.go             (新增 bisync 标签)               │ │
│  │ internal/service/log_cleanup.go     (清理 bisync 目录)               │ │
│  │ internal/service/task_import_export.go (导入/导出 bisync_options)     │ │
│  └──────────┬────────────────────────────────────────────────────────────┘ │
└─────────────┼───────────────────────────────────────────────────────────────┘
              │
┌─────────────▼───────────────────────────────────────────────────────────────┐
│  后端 Runner 层                                                           │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │ internal/runnercli/bisync.go (新建)                                   │ │
│  │  - buildBisyncCommand() - 构建 bisync 命令                             │ │
│  │  - classifyBisyncLogRow() - 解析日志，区分方向                         │ │
│  │  - buildBisyncFinalSummary() - 生成双向总结                           │ │
│  ├───────────────────────────────────────────────────────────────────────┤ │
│  │ internal/runnercli/runner.go (新增 mode=bisync 支持)                  │ │
│  │ internal/runnercli/summary.go (新增 bisync 总结生成)                   │ │
│  └──────────┬────────────────────────────────────────────────────────────┘ │
└─────────────┼───────────────────────────────────────────────────────────────┘
              │
┌─────────────▼───────────────────────────────────────────────────────────────┐
│  数据存储层                                                                │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │ internal/store/models.go    (新增 BisyncOptions, Task.BisyncOptions) │ │
│  │ internal/store/store_tasks.go (SQL 更新，支持 bisync_options)        │ │
│  ├───────────────────────────────────────────────────────────────────────┤ │
│  │ migrations/000009_add_bisync_options.sql (新建)                       │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 四、数据模型设计

### 4.1 Task 模型扩展
**文件**: [internal/store/models.go](/workspace/internal/store/models.go)

```go
type Task struct {
    ID            int64           `json:"id"`
    Name          string          `json:"name"`
    Mode          string          `json:"mode"` // 新增 "bisync"
    SourceRemote  string          `json:"sourceRemote"`
    SourcePath    string          `json:"sourcePath"`
    TargetRemote  string          `json:"targetRemote"`
    TargetPath    string          `json:"targetPath"`
    Options       json.RawMessage `json:"options,omitempty"`
    BisyncOptions json.RawMessage `json:"bisyncOptions,omitempty"` // 新增
    SortOrder     int64           `json:"sortOrder"`
    CreatedAt     time.Time       `json:"createdAt"`
}

type BisyncOptions struct {
    Resync             bool   `json:"resync,omitempty"`
    Compare            string `json:"compare,omitempty"`
    MaxDelete          string `json:"maxDelete,omitempty"`
    CheckAccess        bool   `json:"checkAccess,omitempty"`
    CheckFilename      string `json:"checkFilename,omitempty"`
    ConflictResolve    string `json:"conflictResolve,omitempty"`
    ConflictLoser      string `json:"conflictLoser,omitempty"`
    ConflictSuffix     string `json:"conflictSuffix,omitempty"`
    BackupDir1         string `json:"backupDir1,omitempty"`
    BackupDir2         string `json:"backupDir2,omitempty"`
    CreateEmptySrcDirs bool   `json:"createEmptySrcDirs,omitempty"`
    RemoveEmptyDirs    bool   `json:"removeEmptyDirs,omitempty"`
    Recover            bool   `json:"recover,omitempty"`
}
```

### 4.2 数据库迁移
**文件**: [migrations/000009_add_bisync_options.sql](/workspace/migrations/000009_add_bisync_options.sql) (新建)

```sql
-- 添加 bisync_options 列存储 bisync 特定配置
ALTER TABLE tasks ADD COLUMN bisync_options TEXT;
```

## 五、后端实现设计

### 5.1 RunnerCLI - Bisync 模块（新建）
**文件**: [internal/runnercli/bisync.go](/workspace/internal/runnercli/bisync.go)

**主要功能**:
1. `buildBisyncCommand()` - 构建 rclone bisync 命令及参数
2. `classifyBisyncLogRow()` - 解析 bisync 的 JSON 日志，判断文件传输方向
3. `buildBisyncFinalSummary()` - 生成包含 A→B 和 B→A 两组统计的 final summary

### 5.2 Runner 模块更新
**文件**: [internal/runnercli/runner.go](/workspace/internal/runnercli/runner.go)

**变更**:
- 在 `Start()` 函数中新增 mode == "bisync" 的判断
- 设置 bisync 的工作目录: `/app/data/bisync/<safe-task-name>/`
- 调用 bisync 专用的命令构建和日志解析

### 5.3 Task 服务层更新
**文件**: [internal/service/task.go](/workspace/internal/service/task.go)

**新增函数**:
- `GetBisyncLstFiles(taskID int64) ([]string, error)` - 获取任务的 lst 文件列表
- `DeleteBisyncLstFile(taskID int64, filename string) error` - 删除指定的 lst 文件
- `RollbackBisyncLstFile(taskID int64, filename string) error` - 回滚到指定的 lst 文件版本
- `ResyncBisync(taskID int64) error` - 触发 resync
- `cleanupBisyncDir(taskName string) error` - 清理 bisync 目录（任务删除/模式切换时调用）

**修改函数**:
- `UpdateTask()` - 检测到 mode 从 bisync 切换到其他时，调用清理函数
- `DeleteTask()` - 删除任务前调用清理函数

### 5.4 Tag 服务层更新
**文件**: [internal/service/tag.go](/workspace/internal/service/tag.go)

**变更**:
- 在 `RecalcTags()` 中，将 "bisync" 加入到 action 标签列表

### 5.5 Log Cleanup 服务层更新
**文件**: [internal/service/log_cleanup.go](/workspace/internal/service/log_cleanup.go)

**变更**:
- 在 `cleanup()` 中，除了清理 `/app/data/logs/` 目录，同时也清理 `/app/data/bisync/` 目录下过期的子目录

### 5.6 Task Import/Export 服务层更新
**文件**: [internal/service/task_import_export.go](/workspace/internal/service/task_import_export.go)

**变更**:
- `ExportTasks()` - 在导出的任务对象中加入 bisyncOptions
- `ImportTasks()` - 导入时读取并保存 bisyncOptions

### 5.7 任务 Store 层更新
**文件**: [internal/store/store_tasks.go](/workspace/internal/store/store_tasks.go)

**变更**:
- 所有 SELECT/INSERT/UPDATE 语句中加入 `bisync_options` 字段

### 5.8 任务控制器层更新
**文件**: [internal/controller/task.go](/workspace/internal/controller/task.go)

**新增端点**:
- `GET /api/tasks/{id}/bisync/lst-files` - 获取 lst 文件列表
- `POST /api/tasks/{id}/bisync/delete-lst` - 删除 lst 文件
- `POST /api/tasks/{id}/bisync/rollback-lst` - 回滚到指定的 lst 文件
- `POST /api/tasks/{id}/bisync/resync` - 触发 resync

**文件**: [internal/router/router.go](/workspace/internal/router/router.go)
- 注册上述新端点

## 六、前端实现设计

### 6.1 类型更新
**文件**: [frontend/src/types/index.ts](/workspace/frontend/src/types/index.ts)
**文件**: [frontend/src/components/task/types.ts](/workspace/frontend/src/components/task/types.ts)

新增 `BisyncOptions` 接口，更新 `Task` 和 `CreateForm` 接口以支持 bisyncOptions。

### 6.2 API 层更新
**文件**: [frontend/src/api/task.ts](/workspace/frontend/src/api/task.ts)

新增函数:
- `getBisyncLstFiles(taskId)`
- `deleteBisyncLstFile(taskId, filename)`
- `rollbackBisyncLstFile(taskId, filename)`
- `resyncBisync(taskId)`

### 6.3 Bisync 配置弹窗组件（新建）
**文件**: [frontend/src/components/task/BisyncConfigModal.vue](/workspace/frontend/src/components/task/BisyncConfigModal.vue)

**功能**:
1. Resync 按钮（若为新建任务则禁用，因为没有 taskId）
2. 基本配置区（resync, compare, maxDelete 等）
3. 高级配置区（折叠，包含冲突解决、备份目录等）
4. LST 文件列表（仅编辑已有任务时可用）：
   - 显示文件名、修改时间
   - "回滚" 按钮：点击后将该文件设为当前版本（重命名为标准文件名，如 `path1.lst` / `path2.lst`）
   - "删除" 按钮：删除历史版本

### 6.4 任务表单更新
**文件**: [frontend/src/components/task/AddTaskForm.vue](/workspace/frontend/src/components/task/AddTaskForm.vue)

**变更**:
- 模式下拉新增 "bisync" 选项
- bisync 模式下显示 "配置双向同步" 按钮
- 点击按钮触发 open-bisync-config 事件

### 6.5 任务视图更新
**文件**: [frontend/src/views/TaskView.vue](/workspace/frontend/src/views/TaskView.vue)

**变更**:
- 引入并注册 `BisyncConfigModal` 组件
- 处理弹窗打开/关闭事件
- 任务卡片上仅在任务 mode == "bisync" 时显示配置按钮

### 6.6 运行详情与传输中弹窗更新
**相关文件**:
- [frontend/src/components/task/RunDetailModal.vue](/workspace/frontend/src/components/task/RunDetailModal.vue)
- [frontend/src/components/task/transferring/TransferringModal.vue](/workspace/frontend/src/components/task/transferring/TransferringModal.vue)

**变更**:
- bisync 模式下，分为左右两栏分别显示 A→B 和 B→A 的进度
- final summary 解析与显示也做对应区分

### 6.7 i18n 翻译更新
**文件**: [frontend/src/i18n/zh.ts](/workspace/frontend/src/i18n/zh.ts)
**文件**: [frontend/src/i18n/en.ts](/workspace/frontend/src/i18n/en.ts)

新增 bisync 相关的翻译键。

## 七、关键流程设计

### 7.1 任务创建流程
1. 用户在 AddTaskForm 中选择模式 "bisync"
2. 点击 "配置双向同步" 打开弹窗
3. 选择配置项，点击保存
4. 提交表单，创建任务

### 7.2 任务运行流程
1. 触发任务运行
2. 检测到 mode == "bisync"，设置 bisync 工作目录
3. 检查 lst 文件是否存在，若不存在且用户没有点击 resync，提示用户（？待确认）
4. 执行 rclone bisync 命令
5. 解析日志，区分 A→B 和 B→A 的文件
6. 更新运行状态和进度

### 7.3 模式切换清理流程
1. 用户将任务从 "bisync" 改为 "sync"/"copy"/"move"
2. 后端 UpdateTask 函数检测到前后 mode 不一致
3. 调用 cleanupBisyncDir 函数删除对应目录

### 7.4 日志清理流程
1. LogCleanupService 定时运行
2. 检查 `/app/data/bisync/` 下的子目录
3. 删除超过保留期限的目录

## 八、测试要求

### 8.1 后端测试
- [ ] bisync 命令构建测试
- [ ] bisync 日志解析测试
- [ ] 任务更新/删除时的清理逻辑测试
- [ ] 导入/导出 bisync 任务测试
- [ ] 标签生成测试（包含 bisync）

### 8.2 前端测试
- [ ] BisyncConfigModal 基本渲染测试
- [ ] AddTaskForm bisync 模式显示测试
- [ ] API 调用测试

## 九、约束与注意事项

1. **仅非 CAS 链路可用**: 在 AddTaskForm 中，若检测到源/目标是 CAS 兼容存储，禁用 bisync 选项
2. **目录安全**: 删除 bisync 目录前先检查路径，防止误删
3. **模式兼容性**: bisync 模式不支持单例模式、部分高级选项（？待确认）
4. **版本管理**: rclone 版本需要确保支持 bisync（已升级到 v1.74.2）

## 十、后续迭代方向（可选）

1. 为 bisync 添加全局默认配置（Settings 模块）
2. 优化运行时双向进度显示的实时性

## 十一、实施顺序建议

1. 数据模型与数据库迁移
2. 后端 runner 核心逻辑
3. 后端 service/controller 层
4. 前端组件与 UI
5. 补充各模块（标签、导入导出、日志清理等）
6. 测试
