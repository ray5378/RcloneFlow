# RcloneFlow

[English](./README-en.md) | 中文

## 主要功能

### 📂 存储管理
- 添加、删除 rclone 远程存储
- 支持多种存储类型（S3、Azure、Google Drive 等）
- 查看存储使用情况

### 📁 文件操作
- 浏览远程存储文件
- 复制、移动、重命名文件
- 删除文件/文件夹
- 创建文件夹
- 生成文件公开链接

### 📁 任务管理
- 创建同步任务（复制/移动）
- 设置源路径和目标路径
- 配置传输选项（线程数、带宽限制等）
- 支持 OpenList `.cas` 兼容模式：当目标端已存在 `F.cas` 时，可将其视为源文件 `F` 已满足同步条件

### ⏰ 定时调度
- 支持 Cron 表达式定时触发
- 可设置多个定时任务

### 🔗 Webhook 功能
- 支持 Webhook URL 触发同步（POST 方式）
- 可配置触发来源（手动/定时/Webhook）
- 支持 Webhook POST 通知（任务完成时主动推送）
- 支持企业微信、通用 Markdown 格式

### 📊 实时进度
- 传输过程实时显示百分比、速度、预估时间
- 进度条清晰展示传输状态


### 📜 历史记录
- 记录每次同步的详细结果
- 查看传输文件列表、成功/失败状态
- 支持按时间清理历史

## 快速开始

### Docker Compose 部署

```yaml
services:
  rcloneflow:
    image: ray5378/rcloneflow:latest
    container_name: rcloneflow
    network_mode: host
    environment:
      - TZ=Asia/Shanghai
      - APP_ADDR=:17870
      - APP_DATA_DIR=/app/data
      - RCLONE_CONFIG=/app/data/rclone.conf
      - EMBED_RC=true
      - RCLONE_RC_URL=http://127.0.0.1:5572
      - RCLONE_RC_USER=rc
      - RCLONE_RC_PASS=rcpass
      - PUID=${PUID:-1000}
      - PGID=${PGID:-1000}
    volumes:
      - ./app/data:/app/data
      - ./webdav-cache:/webdav-cache  # WebDAV VFS 缓存目录：访问的文件会完整下载到此目录。如果担心磨损 SSD，可将此路径映射到机械硬盘目录
    restart: always
networks: {}
```

### 配置 rclone

将你的 rclone 配置文件放到 `./app/data/rclone.conf`

### 访问界面

打开浏览器访问 `http://<服务器IP>:17870`

## 界面说明

### 任务列表
- 显示所有同步任务
- 点击任务卡片可查看详情和实时进度
- 支持手动运行、编辑、删除任务

### 历史记录
- 查看每次同步的详细结果
- 包括传输时间、文件数量、传输大小等

### 设置
- 配置默认传输选项
- 设置 Webhook 通知
- 配置历史记录保留天数

## 注意事项

- 首次使用请先在「远程存储」中添加 rclone 远程存储
- Webhook 触发需要在任务设置中开启并配置 URL
- 建议定期清理历史记录以节省存储空间
