# RcloneFlow 开发文档

本文档面向开发者和运维人员，包含技术细节、API 说明和部署指南。

## 开发规范

### 前端规范

#### UI 组件
- **弹窗必须使用自定义 Vue Modal 组件**
- **禁止使用浏览器原生 `alert()`、`confirm()`、`prompt()`**
- 所有用户提示必须使用自定义 Toast 或 Modal

#### 组件引用方式
```vue
<!-- Modal 弹窗 -->
<div v-if="showModal" class="modal-overlay" @click.self="showModal=false">
  <div class="modal-content">
    ...
  </div>
</div>
```

#### 样式类名
- `modal-overlay` - 遮罩层
- `modal-content` - 弹窗内容区
- `modal-header` - 标题栏
- `modal-body` - 主体内容
- `modal-footer` - 底部按钮区
- `close-btn` - 关闭按钮
- `detail-item` - 表单项
- `trigger-opt` - 复选框选项样式

## 技术架构

### 核心组件
- **前端**: Vue 3 + TypeScript + Vite
- **后端**: Go + Gin 框架
- **数据库**: SQLite
- **传输引擎**: rclone (CLI + RC 双模式)

### 核心概念

#### 稳态进度 (Stable Progress)
运行中的任务使用"稳态进度"展示，包括阶段信息：
- `preparing` - 准备阶段
- `transferring` - 传输中
- `between_files` - 文件间隔
- `finalizing` - 完成中

#### 历史冻结 (Summary Freeze)
任务结束时，将最终总结一次性写入数据库（`summary.finalSummary`），历史页面只渲染这份冻结数据，不再更新。

#### JWT 鉴权
日志下载和所有 API 需携带有效 Token。

#### 单例模式 (Singleton Mode)
开启单例模式的任务在触发时会检查是否有其他任务正在运行：
- 如果有其他任务在运行，跳过本次执行
- 如果没有其他任务在运行，启动当前任务
- 使用数据库事务确保原子性

## API 接口

### 基础信息
- 基础路径: `/api`
- 认证方式: JWT Bearer Token

### 主要接口

#### 任务相关
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/tasks | 获取任务列表 |
| POST | /api/tasks | 创建任务 |
| PATCH | /api/tasks | 更新任务 |
| DELETE | /api/tasks/{id} | 删除任务 |
| POST | /api/tasks/{id}/run | 运行任务 |

#### 运行记录
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/runs | 历史列表 |
| GET | /api/runs/active | 运行中的任务 |
| GET | /api/runs/{id} | 单条记录 |
| GET | /api/runs/{id}/files | 文件列表 |

#### 远程存储
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/remotes | 存储列表 |
| POST | /api/remotes | 添加存储 |
| GET | /api/config | rclone 配置 |
| GET | /api/providers | 支持的存储类型 |

## 构建与部署

### 构建前端
```bash
cd frontend
npm install
npm run build
```

前端构建产物在 `../web` 目录。

### 构建 Docker 镜像
```bash
docker build --no-cache -t ray5378/rcloneflow:dev .
```

### 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| APP_ADDR | :17870 | 服务地址 |
| APP_DATA_DIR | /app/data | 数据目录 |
| LOG_LEVEL | info | 日志级别 |
| PRECHECK_MODE | none | 预检模式 |
| PROGRESS_FLUSH_INTERVAL | 5s | 进度刷新间隔 |
| FINISH_WAIT_INTERVAL | 5s | 完成后等待间隔 |
| FINAL_SUMMARY_RETENTION_DAYS | 7 | 历史保留天数 |
| CLEANUP_INTERVAL_HOURS | 24 | 清理扫描间隔 |

## 数据库

### 主要表结构

#### tasks - 任务表
- id, name, mode, source_remote, source_path
- target_remote, target_path, options (JSON)
- schedule (Cron), webhook_url

#### runs - 运行记录表
- id, task_id, status, trigger
- start_at, finished_at
- summary (JSON) - 包含 finalSummary

### 清理机制
- `FINAL_SUMMARY_RETENTION_DAYS`: finalSummary 保留天数
- `CLEANUP_INTERVAL_HOURS`: 清理扫描间隔

## Webhook 通知

### 配置
- 设置 Webhook URL
- 选择触发条件（成功/失败）
- 选择触发来源（手动/定时/Webhook）

### 载荷格式
```json
{
  "taskName": "备份任务",
  "status": "success",
  "startAt": "2024-01-01T10:00:00Z",
  "finishedAt": "2024-01-01T10:05:00Z",
  "stats": {
    "transferredBytes": 1024000,
    "totalBytes": 2048000,
    "avgSpeedBps": 341333,
    "filesCopied": 10,
    "filesFailed": 0
  }
}
```

## 调试

### 常见问题

#### 1. 页面显示旧前端
原因：挂载覆盖了 `/app/web`
解决：不要挂载覆盖 /app/web 目录

#### 2. 镜像无 /bin/sh
解决：使用 `docker create` 创建容器后用 `docker cp` 抽取文件

#### 3. rclone 操作失败
检查 rclone.conf 配置是否正确，确保远程存储名称与任务中一致

### 查看日志
```bash
docker logs rcloneflow
```

### 数据库位置
挂载卷中的 `data` 目录下可找到 SQLite 数据库文件。

## 开发指南

### 前端开发
```bash
cd frontend
npm install
npm run dev    # 开发模式
npm run build # 生产构建
```

### 后端开发
```bash
go build -o server ./cmd/server
./server
```

### 代码结构
```
rcloneflow/
├── cmd/server/      # 入口
├── internal/
│   ├── app/         # 应用初始化
│   ├── controller/  # 控制器
│   ├── router/      # 路由
│   ├── service/     # 业务逻辑
│   ├── store/       # 数据层
│   └── adapter/     # 适配器
├── frontend/        # 前端
├── web/             # 前端构建产物
└── Dockerfile
```
