# RcloneFlow

A powerful web-based Rclone management interface for multi-storage copy/sync/move task management.

[中文文档](README_ZH.md)

## Features

- **Multi-Storage Management** - Add, edit, and manage multiple Rclone remotes
- **File Browser** - Browse and navigate remote storage files
- **Task Management** - Create and manage copy/sync tasks between remotes
- **Scheduled Tasks** - Set up automated sync with cron-like scheduling
- **Run History** - Track task execution history and status
- **Modern UI** - Clean, responsive web interface

## Requirements

- Go 1.22+
- Rclone (with RC mode enabled)
- Git

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/ray5378/RcloneFlow.git
cd RcloneFlow
```

### 2. Configure Rclone

Make sure Rclone is installed and configured with your remotes. You can configure it at `~/.config/rclone/rclone.conf` or set the `RCLONE_CONFIG` environment variable.

### 3. Start Rclone RC Server

```bash
rclone rcd --rc-user=your_user --rc-pass=your_pass --rc-addr=localhost:5572
```

Or use environment variables:
```bash
export RCLONE_RC_URL=http://localhost:5572
export RCLONE_RC_USER=your_user
export RCLONE_RC_PASS=your_pass
```

### 4. Build and Run

```bash
# Build
go build -o server ./cmd/server

# Run
./server
```

The server will start on port 17870 by default. Access it at http://localhost:17870

### 5. Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_ADDR` | Server address | `:17870` |
| `APP_DATA_DIR` | Data directory | `./data` |
| `RCLONE_RC_URL` | Rclone RC URL | `http://127.0.0.1:5572` |
| `RCLONE_RC_USER` | Rclone RC username | - |
| `RCLONE_RC_PASS` | Rclone RC password | - |
| `RCLONE_RC_TIMEOUT` | RC timeout | `120s` |

## Docker

### Build Image

```bash
docker build -t rcloneflow .
```

### Run Container

```bash
docker run -d \
  --name rcloneflow \
  -p 17870:17870 \
  -e RCLONE_RC_URL=http://host.docker.internal:5572 \
  -e RCLONE_RC_USER=your_user \
  -e RCLONE_RC_PASS=your_pass \
  -v /path/to/rclone/config:/root/.config/rclone \
  rcloneflow
```

### Docker Compose

```yaml
version: '3.8'
services:
  rcloneflow:
    build: .
    ports:
      - "17870:17870"
    environment:
      - RCLONE_RC_URL=http://rclone:5572
      - RCLONE_RC_USER=your_user
      - RCLONE_RC_PASS=your_pass
    volumes:
      - ./data:/app/data
      - /path/to/rclone/config:/root/.config/rclone
    depends_on:
      - rclone

  rclone:
    image: rclone/rclone
    container_name: rclone
    volumes:
      - /path/to/rclone/config:/config/rclone
      - /path/to/your/data:/data
    command: rcd --rc-user=your_user --rc-pass=your_pass --rc-addr=0.0.0.0:5572
```

## Project Structure

```
RcloneFlow/
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── app/            # HTTP server and API handlers
│   ├── rclone/         # Rclone RC client wrapper
│   ├── scheduler/      # Task scheduling logic
│   └── store/          # Data persistence (SQLite)
├── web/
│   ├── index.html      # Frontend SPA (Vue.js)
│   └── vendor/         # Vue.js CDN bundle
├── data/               # Application data directory
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/remotes` | List all remotes |
| POST | `/api/remotes` | Create new remote |
| PUT | `/api/remotes` | Update remote config |
| GET | `/api/remotes/config/{name}` | Get remote config |
| POST | `/api/remotes/test` | Test remote connection |
| GET | `/api/providers` | List supported providers |
| GET | `/api/browser/list` | List directory contents |
| GET/POST | `/api/tasks` | List/Create tasks |
| POST | `/api/tasks/{id}/run` | Run a task |
| GET/POST | `/api/schedules` | List/Create schedules |
| GET | `/api/runs` | List run history |

## Development

### Build Frontend (Optional)

The frontend is bundled as a single HTML file with embedded Vue.js. For development:

```bash
# Frontend is served directly from web/index.html
# No build step required for basic development
```

### Run Tests

```bash
go test ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details.

## Acknowledgments

- [Rclone](https://rclone.org/) - The powerful cloud storage sync tool
- [Vue.js](https://vuejs.org/) - The progressive JavaScript framework
