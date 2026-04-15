# RcloneFlow

[English](./README-en.md) | [中文](./README.md)

RcloneFlow is a file sync management tool with scheduling, Webhook triggers, real-time progress tracking, and history management.

## Features

### 📂 Storage Management
- Add/remove rclone remotes
- Support various storage types (S3, Azure, Google Drive, etc.)
- View storage usage

### 📁 File Operations
- Browse remote files
- Copy, move, rename files
- Delete files/folders
- Create folders
- Generate public sharing links

### 📁 Task Management
- Create sync tasks (copy/move)
- Set source and target paths
- Configure transfer options (threads, bandwidth limits, etc.)

### ⏰ Scheduled Sync
- Cron expression support for timed triggers
- Multiple scheduled tasks supported

### 🔗 Webhook Features
- Trigger sync via Webhook URL (POST)
- Configure trigger sources (manual/scheduled/webhook)
- Webhook POST notifications (push when task completes)
- WeCom and Markdown format support

### 📊 Real-time Progress
- Live percentage, speed, and ETA display
- Clear progress bar showing transfer status

### 📜 History
- Detailed results for each sync
- View file lists, success/failure status
- Auto-cleanup by age

### 🔔 Webhook Notifications
- Notifications when sync completes
- WeCom and Markdown format support
- Configurable notification content (stats, file list)

## Quick Start

### Docker Compose Deployment

```yaml
services:
  rcloneflow:
    image: ray5378/rcloneflow:latest
    platform: linux/amd64
    container_name: rcloneflow
    environment:
      - TZ=Asia/Shanghai
      - APP_ADDR=:17870
      - APP_DATA_DIR=/app/data
      - RCLONE_CONFIG=/app/data/rclone.conf
      # Embedded RC (for remotes/providers/config/browser)
      - EMBED_RC=true
      - RCLONE_RC_URL=http://127.0.0.1:5572
      - RCLONE_RC_USER=rc
      - RCLONE_RC_PASS=rcpass
      # Log level: debug|info|warn|error
      - LOG_LEVEL=info
    volumes:
      - ./data:/app/data
    ports:
      - "17870:17870"
    restart: always
```

### Configure rclone

Place your rclone config file at `./data/rclone.conf`

### Access UI

Open browser: `http://<serverIP>:17870`

### Default Credentials

- Username: `admin`
- Password: `admin`

## Interface

### Task List
- View all sync tasks
- Click task card for details and live progress
- Manual run, edit, delete tasks

### History
- View sync results
- Transfer time, file count, size, etc.

### Settings
- Configure default transfer options
- Set up Webhook notifications
- Configure history retention

## Notes

- Add rclone remotes in "Remote Storage" before creating tasks
- Enable and configure Webhook URL in task settings to use webhook triggers
- Regular history cleanup recommended to save storage space
