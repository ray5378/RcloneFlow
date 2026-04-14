# RcloneFlow

[English README](./README.en.md)

RcloneFlow = Rclone + Web UI + 调度 + 历史。

本分支重点：稳态进度、JWT 鉴权 API、历史冻结模型；浏览=RC，文件操作=CLI。

## 核心概念
- 稳态进度（preparing/transferring/between_files/finalizing）用于“运行中”展示
- 历史冻结：任务结束时一次性写入 summary.finalSummary，历史页面只渲染该最终态
- JWT 鉴权：日志下载与所有 API 需携带 Token

## 历史冻结（summary.finalSummary）
任务结束（成功/失败）时，后端把最终总结一次性写入 DB；前端历史仅展示这份冻结数据。

字段示例：
- startAt / finishedAt（RFC3339）
- durationSec / durationText（如“1小时2分3秒”）
- result（success/failed）
- transferredBytes / totalBytes / avgSpeedBps
- counts { copied, deleted, skipped, failed, total }
- files: { path, action, status, sizeBytes?, at? } 列表（按保留期清理，不默认截断）

## 接口速览
- GET /api/runs：历史列表（含 durationSeconds/durationText 与 summary.finalSummary）
- GET /api/runs/{id}：单个历史
- GET /api/runs/active：仅“运行中”稳态；runRecord 含持续时长，stableProgress 含阶段/百分比/速度
- GET /api/runs/{id}/files：原始/解析日志（历史 UI 不依赖此接口）

## 保留与清理
- FINAL_SUMMARY_RETENTION_DAYS：finalSummary 与历史保留天数（默认 7）
- CLEANUP_INTERVAL_HOURS：清理扫描间隔（默认 24 小时）
- 日志保留独立：LOG_RETENTION_DAYS / LOG_CLEANUP_INTERVAL_HOURS（如有配置）

## 设置（/api/settings + APP_DATA_DIR/settings.json）
优先级：环境变量 > settings.json > 内置默认；UI 在“设置 → 默认设置”中可视化管理。

常用键：
- ACCESS_TOKEN_TTL：如 24h
- REFRESH_TOKEN_TTL：如 90d
- PRECHECK_MODE：none | size
- PROGRESS_FLUSH_INTERVAL：如 5s
- PROGRESS_FLUSH_MIN_DELTA_PCT：0-100（可小数）
- PROGRESS_FLUSH_MIN_DELTA_BYTES：字节（UI 以 MB 显示，保存时换算）
- FINISH_WAIT_INTERVAL：如 5s
- FINISH_WAIT_TIMEOUT：如 5h
- FINAL_SUMMARY_RETENTION_DAYS / CLEANUP_INTERVAL_HOURS
- WEBHOOK_MAX_FILES：Webhook 文件名数量上限（0=不限制，默认 0）

### Webhook 通知
- 载荷包含任务/运行/统计摘要与文件名（最多 N 个，0 表示全部）；WeCom/通用 Markdown 双端点
- 提示：文件极多时接收端可能受请求体限制（413/超时），可设置合适上限或采用分批/摘要

## 运行与构建
- 构建镜像：
  - docker build --no-cache -t ray5378/rcloneflow:next-master -f rcloneflow/Dockerfile rcloneflow
- 前端构建：Vite outDir=../web，最终镜像复制到 /app/web
- 注意：不要挂载覆盖 /app/web（否则会看到旧前端）；镜像极简，无 /bin/sh，调试请使用 docker create+cp 抽取 /app/web

## Docker Compose 示例
```yaml
version: '3.8'
services:
  rcloneflow:
    image: ray5378/rcloneflow:next-master
    container_name: rcloneflow
    ports:
      - "17871:17871"   # API/UI（示例，实际以镜像内暴露端口为准）
      - "17872:17872"
      - "17873:17873"
    environment:
      - APP_ADDR=:17871
      - APP_DATA_DIR=/app/data
      - LOG_LEVEL=info
      - PRECHECK_MODE=none
      - PROGRESS_FLUSH_INTERVAL=5s
      - PROGRESS_FLUSH_MIN_DELTA_PCT=1
      - PROGRESS_FLUSH_MIN_DELTA_BYTES=52428800
      - FINISH_WAIT_INTERVAL=5s
      - FINISH_WAIT_TIMEOUT=5h
      - WEBHOOK_MAX_FILES=0
    volumes:
      - ./data:/app/data   # rclone.conf / settings.json / DB / 日志
    restart: unless-stopped
```

访问：
- UI 示例：http://<宿主机IP>:17871 或 17873（按你的映射）
- APP_DATA_DIR=/app/data：
  - /app/data/rclone.conf：rclone 配置
  - /app/data/settings.json：系统设置
  - /app/data/…：历史/日志/数据库

## 前端行为（中文）
- 历史页只渲染 finalSummary：开始/结束/固定时长/结果/均速/体量/文件
- 运行中显示“稳态进度”（阶段/百分比/速度），不将实时帧写入 DB
- 任务卡片细节：
  - 深色模式进度条高反差（白色填充）
  - 进度达 100% 后等 20s 自动刷新两次（+1s 再拉一次），并保留 20s 可视停留；若 25s 以上完全无变化将自检强制刷新
  - “最后 1 个文件”显示收敛：接近完成时 UI 钳制为 100% 并对齐已传数量

## 浏览与文件操作（固定方案）
- 浏览（RC）：/api/browser/list → rclone RC（ListPath）；在 SMB/WebDAV 等后端更稳
- 文件操作（CLI）：/api/fs/* → rclone CLI（复制/移动/重命名/删除/新建/公开链接）
  - WebDAV MOVE 失败自动回退为 copy(+dir)+delete/purge，并做目标可见性与源端消失等待
  - SMB 遇 share 解析异常时按需去重首段 share 后重试
  - 路径净化：只移除路径段“末尾 ASCII 冒号”，保留中间冒号与全角“：”
- 存储管理（RC 保留）：/api/remotes、/api/config、/api/providers、/api/usage/{fs}、/api/fsinfo/{fs}

## 注意事项
- 不要挂载覆盖 /app/web（会导致旧前端）
- 镜像无 /bin/sh；调试请使用 docker create+cp 抽取前端包
- 浏览=RC、文件操作=CLI 为固定方案；不再提供 RC/CLI 切换项
