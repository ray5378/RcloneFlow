# RcloneFlow

基于 Web 的 Rclone 管理界面（Go + Vue 3）。支持创建/运行/调度任务、实时进度与日志、统一“传输选项”配置、JWT 登录。提供官方 Docker 镜像与 docker-compose 示例。

[English](README_EN.md)

## 最近变更（要点）
- 镜像与构建
  - 基础镜像切换为 Alpine，支持 linux/amd64；前端采用多阶段构建（Vite → /app/web）。
  - 内置 rclone v1.73.4（构建阶段从 downloads.rclone.org 下载）。
  - 仅发布 master 与 master-<sha> 标签；支持 --no-cache 重建。
  - 健康检查内置 wget，请求首页 /（start-period 45s），容器更稳（不要在 compose 里用 curl /healthz）。
- 运行与解析
  - 统一 CLI Runner，命令行仅使用 --stats-one-line（兼容 v1.73.x）。
  - 实时进度解析增强：同时解析 stdout/stderr，容忍 KiB/s/MiB/s、mm:ss/hh:mm:ss/"-" 等格式；多段拼接取最后一段。
  - /api/runs/active 返回扁平实时字段：bytes/totalBytes/speed/eta/percentage（realtimeStatus 与 derivedProgress 均可读）。
  - 仅保留单一日志文件 run-<id>-stderr.log（stdout 合并写入）；下载统一 /api/runs/{id}/log。
- 传输选项（可视化）
  - 设置页与任务卡片提供“传输选项”弹窗（全局/任务级覆盖，中文）。
  - buffer-size 智能补单位：数字或纯数字字符串自动补 M（例如 16 → 16M）。
  - 不再默认拼接 --timeout；仅在你配置 connTimeout/expectContinueTimeout 时才拼接对应参数。
- 历史与清理
  - 历史详情页与实时卡片显示一致的速度/进度/已传/总大小。
  - 新增“历史与日志保留”设置（默认 7 天）：自动清理过期 run 记录与 /app/data/logs/run-*-stderr.log。

## 快速开始（Docker 推荐）

### docker run（最短路径）
```bash
# 数据目录（确保对 UID 1000 可写）
mkdir -p ./data

# 运行（仅 amd64）
export DOCKER_DEFAULT_PLATFORM=linux/amd64
docker run -d --name rcloneflow \
  -p 17870:17870 \
  -e TZ=Asia/Shanghai \
  -e APP_ADDR=:17870 \
  -e APP_DATA_DIR=/app/data \
  -e RCLONE_CONFIG=/app/data/rclone.conf \
  -e EMBED_RC=true \
  -e RCLONE_RC_URL=http://127.0.0.1:5572 \
  -e RCLONE_RC_USER=rc \
  -e RCLONE_RC_PASS=rcpass \
  -v $(pwd)/data:/app/data \
  ray5378/rcloneflow:master
```
访问 http://你的地址:17870，默认 JWT 登录（首次登录后请修改密码）。

### docker-compose（建议）
```yaml
version: "3.8"
services:
  rcloneflow:
    image: ray5378/rcloneflow:master
    platform: linux/amd64
    container_name: rcloneflow
    environment:
      - TZ=Asia/Shanghai
      - APP_ADDR=:17870
      - APP_DATA_DIR=/app/data
      - RCLONE_CONFIG=/app/data/rclone.conf
      # 可选嵌入 RC（用于 remotes/providers/browser）
      - EMBED_RC=true
      - RCLONE_RC_URL=http://127.0.0.1:5572
      - RCLONE_RC_USER=rc
      - RCLONE_RC_PASS=rcpass
    volumes:
      - ./data:/app/data
    ports:
      - 17870:17870
    restart: unless-stopped
```

> 注意：请不要在 compose 中使用 curl /healthz 作为检查；镜像内置 wget + 首页 / 更稳。

## 主要功能
- 多存储管理（SMB/WebDAV 等），文件浏览与基础操作
- 任务管理：复制/同步/移动，支持定时任务（cron 风格）
- JWT 登录与权限校验
- 统一“传输选项”配置（全局/任务级）：
  - 常用项：transfers、checkers、buffer-size、include/exclude 等
  - 超时：仅在你设置时拼接 --contimeout/--expect-continue-timeout；默认不拼接 --timeout
  - buffer-size 智能补单位（16 → 16M；"64Mi" 原样透传）
- 实时进度：bytes/totalBytes/speed/eta/percentage（卡片/全局/历史详情一致）
- 日志：仅 run-<id>-stderr.log；首行包含完整命令 + effectiveOptions，下载接口统一
- 历史与日志保留：默认 7 天，可在设置修改；自动清理过期数据

## API 速览（新增/调整）
- 实时运行：`GET /api/runs/active`
  - 兼容旧结构，并提供扁平实时字段：realtimeStatus.bytes/totalBytes/speed/eta/percentage
- 传输选项：
  - 全局：`GET/PUT /api/settings/transfer`（持久化到 APP_DATA_DIR/transfer_settings.json）
  - 任务：`PATCH /api/tasks`（仅更新 Options）
- 历史与日志保留：`GET/PUT /api/settings/housekeeping`（runRetentionDays/logRetentionDays）
- 日志下载：`GET /api/runs/{id}/log`（仅 stderr 单文件）

## 环境变量（常用）
| 变量 | 说明 | 默认 |
|------|------|------|
| APP_ADDR | 服务监听地址 | :17870 |
| APP_DATA_DIR | 数据目录 | /app/data |
| RCLONE_CONFIG | rclone.conf 路径 | /app/data/rclone.conf |
| EMBED_RC | 是否嵌入 RC | true |
| RCLONE_RC_URL/USER/PASS | RC 地址与认证 | - |
| LOG_LEVEL | 日志级别 | info |
| LOG_RETENTION_DAYS | 日志保留天数 | 7 |
| LOG_CLEANUP_INTERVAL_HOURS | 日志清理周期（小时） | 24 |
| RCLONE_USE_JSON_LOG | 启用 rclone JSON 日志 | false |

## 构建镜像（本地）
```bash
export DOCKER_DEFAULT_PLATFORM=linux/amd64
SHA=$(git rev-parse --short HEAD)
docker build --no-cache \
  -t ray5378/rcloneflow:master \
  -t ray5378/rcloneflow:master-$SHA .
```

## Webhook 触发（外部调用）
- 端点（公开，无需认证）
  - 按任务ID：GET/POST /webhook/{taskId}
  - 按自定义ID：GET/POST /webhook/{customId}（在任务卡片“Webhook”按钮设置 webhookId）
- 返回示例：`{"started": true, "taskId": 12}`
- 示例
```bash
curl -X POST http://localhost:17870/webhook/12
curl http://localhost:17870/webhook/gate-front-01
```
- 注意
  - 该端点无需认证；若暴露到公网，建议通过反向代理限制来源或路径前缀。

## 迁移提示（老版本 → 本版）
- 健康检查：改用 wget 请求首页 /；若 compose 里写了 curl /healthz，请改/删除。
- 日志：仅保留 stderr.log；下载接口统一到 /api/runs/{id}/log。
- 进度：仅 --stats-one-line；不再使用 --stats-one-line-json。
- 超时：不再默认 --timeout；仅当你在“传输选项”里设置了 connTimeout/expectContinueTimeout 才拼接对应参数。
- buffer-size：数字或纯数字字符串自动补 M；已带单位字符串原样保留。

## 项目结构（简要）
```
internal/
  adapter/      # CLI/RC 适配
  app/          # Runner flags 映射等
  controller/   # /api 控制器（含 /api/runs/active, /api/settings/transfer 等）
  router/       # 路由
  scheduler/    # 定时/轮询
  service/      # 业务
  store/        # SQLite
frontend/
  src/components/TransferOptions.vue  # 传输选项弹窗（全局/任务）
  src/views/TaskView.vue              # 任务/历史视图（含命令行模式与下载日志按钮）
web/                                   # 生产静态资源（由 Vite 构建）
```

## 常见问题（FAQ）
- 容器 Unhealthy？
  - 请不要在 compose 里使用 curl /healthz；改为 wget + 首页 /，或直接删掉 compose 的 healthcheck 使用镜像内置检查。
- 实时进度一直为 0？
  - 使用 v1.73.x 的 --stats-one-line：我们已同时解析 stdout/stderr，并兼容 KiB/s/MiB/s/ETA 多种格式。如仍为 0，请提供一行日志样例。
- buffer-size 仍然显示 16？
  - 新版已做两道兜底（映射层 + 启动前二次校验），重建容器后会输出 16M。

## 许可证
MIT License
