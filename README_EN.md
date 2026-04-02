# RcloneFlow

A web-based Rclone management interface for multi-storage copy/sync/move task management.

[中文](README.md)

## Features

- **Multi-storage Management** - Add, edit and manage multiple Rclone remotes
- **File Browser** - Browse and navigate remote storage files with clipboard operations
- **Task Management** - Create and manage copy/sync/move tasks between storages
- **Scheduled Tasks** - Automate sync with cron-style scheduling
- **Run History** - Track task execution history and real-time status
- **Real-time Status Sync** - Background sync with rclone job API
- **Modern UI** - Clean, responsive web interface
- **Unified Error Handling** - Toast notifications, friendly error messages

## Requirements

- Go 1.22+
- Rclone (RC mode enabled)
- Git

## Quick Start

### 1. Clone the repository

```bash
git clone https://github.com/ray5378/RcloneFlow.git
cd RcloneFlow
```

### 2. Configure Rclone

Make sure Rclone is installed and remotes are configured. Config is usually at `~/.config/rclone/rclone.conf`.

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
# Build backend
go build -o server ./cmd/server

# Run
./server
```

Server starts on port 17870 by default. Visit http://localhost:17870

### 5. Configuration

Config file `config.yaml`:
```yaml
server:
  addr: ":17870"
  static_dir: "./web"

rclone:
  rc_url: "http://127.0.0.1:5572"
  rc_user: ""
  rc_pass: ""
  timeout: "120s"

storage:
  data_dir: "./data"

log:
  level: "info"
  output: "stdout"

sync:
  pool_interval: 5      # Job status sync interval (seconds)
  schedule_interval: 1  # Schedule check interval (minutes)
```

### 6. Environment Variables

| Variable | Description | Default |
|------|------|--------|
| `APP_ADDR` | Server address | `:17870` |
| `APP_DATA_DIR` | Data directory | `./data` |
| `RCLONE_RC_URL` | Rclone RC URL | `http://127.0.0.1:5572` |
| `RCLONE_RC_USER` | Rclone RC user | - |
| `RCLONE_RC_PASS` | Rclone RC password | - |
| `RCLONE_RC_TIMEOUT` | RC timeout | `120s` |

## Project Structure

```
RcloneFlow/
├── cmd/
│   └── server/              # Main application entry
├── internal/
│   ├── adapter/            # Rclone API adapter layer
│   ├── controller/         # HTTP controllers
│   ├── dao/                # Data Access Layer
│   ├── service/            # Business Logic Layer
│   ├── scheduler/          # Task scheduler
│   ├── router/             # Route definitions
│   ├── store/              # Database wrapper (SQLite)
│   └── config/             # Configuration management
├── frontend/               # Frontend source (Vue 3 + TypeScript)
│   └── src/
│       ├── api/            # API layer (unified封装)
│       ├── components/     # Vue components
│       └── views/          # Page views
├── migrations/              # Database migrations (goose)
├── web/                    # Compiled frontend files
├── config.yaml             # Configuration file
├── Dockerfile
└── docker-compose.yml
```

## Architecture

### Backend (Go)

- **Router** - HTTP routing, request dispatch
- **Controller** - Parameter validation, service invocation
- **Service** - Business logic
- **DAO** - Database operations
- **Adapter** - Rclone API adapter

### Frontend (Vue 3 + TypeScript)

- **API Layer** - Unified API calls
  - `api/client.ts` - HTTP client with interceptors
  - `api/errors.ts` - Unified error handling
  - `api/task.ts` - Task APIs
  - `api/run.ts` - Run history APIs
  - `api/remote.ts` - Remote storage APIs
  - `api/browser.ts` - File browser APIs
- **Components** - Toast, Modal etc.
- **Views** - TaskView, BrowserView etc.

## API Endpoints

### Tasks
| Method | Endpoint | Description |
|------|------|------|
| GET | `/api/tasks` | List all tasks |
| POST | `/api/tasks` | Create task |
| PUT | `/api/tasks` | Update task |
| DELETE | `/api/tasks/{id}` | Delete task |
| POST | `/api/tasks/{id}/run` | Run task |

### Schedules
| Method | Endpoint | Description |
|------|------|------|
| GET | `/api/schedules` | List all schedules |
| POST | `/api/schedules` | Create schedule |
| PUT | `/api/schedules/{id}` | Enable/disable schedule |
| DELETE | `/api/schedules/{id}` | Delete schedule |

#### Schedule Spec Format
Schedule uses standard 5-field cron expression with `|` as field separator.

**Format:** `minute|hour|day|month|week`

| Field | Description | Example |
|------|------|------|
| minute | Minute | `*` every minute, `00,30` specific minutes |
| hour | Hour | `*` every hour, `17,19` specific hours |
| day | Day of month | `*` every day, `1,15` specific days |
| month | Month | `*` every month, `1,3,5` specific months |
| week | Weekday | `*` every day, `1,3,5` Mon,Wed,Fri |

Examples:
- `43|17,19|*|*|*` = 43 min past 17:00 or 19:00 every day → cron: `43 17,19 * * *`
- `00|09|*|*|1,3,5` = 09:00 every Mon,Wed,Fri → cron: `0 00 09 * * 1,3,5`
- `30|14|15|*|*` = 14:30 on 15th of every month → cron: `0 30 14 15 * *`

### Runs
| Method | Endpoint | Description |
|------|------|------|
| GET | `/api/runs` | List run history |
| GET | `/api/runs/{id}` | Get run details |
| DELETE | `/api/runs/{id}` | Clear run record |
| GET | `/api/runs/active` | Get active runs with real-time status |

### Remotes
| Method | Endpoint | Description |
|------|------|------|
| GET | `/api/remotes` | List all remotes |
| POST | `/api/remotes` | Create remote |
| PUT | `/api/remotes` | Update remote |
| GET | `/api/remotes/config/{name}` | Get remote config |
| DELETE | `/api/config/{name}` | Delete remote |
| POST | `/api/remotes/test` | Test remote connection |
| GET | `/api/providers` | List supported remote types |

### Browser
| Method | Endpoint | Description |
|------|------|------|
| GET | `/api/browser/list` | List directory contents |
| POST | `/api/fs/copy` | Copy file |
| POST | `/api/fs/move` | Move file |
| POST | `/api/fs/copyDir` | Copy directory |
| POST | `/api/fs/moveDir` | Move directory |
| POST | `/api/fs/delete` | Delete file |
| POST | `/api/fs/purge` | Delete directory |
| POST | `/api/fs/mkdir` | Create directory |

## Database Migrations

Using goose for versioned database migrations:

```bash
# Check migration status
goose status

# Run migrations
goose up

# Rollback one version
goose down

# Create new migration
goose create add_new_field
```

Migration files are in `migrations/` directory.

## Development

### Frontend Development

```bash
cd frontend

# Install dependencies
npm install

# Dev mode (hot reload)
npm run dev

# Run tests
npm test

# Coverage report
npm run test:coverage

# Production build
npm run build
```

### Backend Development

```bash
# Run tests
go test ./...

# Coverage report
go test -cover ./...

# Build
go build -o server ./cmd/server
```

## Docker Deployment

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

## Contributing

Pull requests are welcome!

## License

MIT License - See LICENSE file.

## Acknowledgements

- [Rclone](https://rclone.org/) - Powerful cloud storage sync tool
- [Vue.js](https://vuejs.org/) - Progressive JavaScript framework
