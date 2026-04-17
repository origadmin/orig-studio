# Administrator Documentation

## Table of Contents

- [1. Welcome](#1-welcome)
- [2. Installation](#2-installation)
- [3. Configuration](#3-configuration)
- [4. User Management](#4-user-management)
- [5. Media Management](#5-media-management)
- [6. Category and Tag Management](#6-category-and-tag-management)
- [7. Dashboard](#7-dashboard)
- [8. Transcoding](#8-transcoding)
- [9. Maintenance](#9-maintenance)

---

## 1. Welcome

This document is for OrigCMS administrators responsible for setting up the system, maintaining operations, and making modifications.

### System Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | 2 cores | 4+ cores |
| Memory | 2GB | 8GB+ |
| Storage | 50GB | 500GB+ |
| OS | Ubuntu 22.04 | Ubuntu 22.04 LTS |

---

## 2. Installation

### 2.1 Prerequisites

- Go 1.21+
- Node.js 18+
- PostgreSQL 15+
- Redis
- FFmpeg

### 2.2 Docker Installation

```bash
# Clone repository
git clone https://github.com/origadmin/orig-cms.git
cd orig-cms

# Build and run
docker compose up
```

The system will be available at `http://localhost` with default admin credentials.

### 2.3 Manual Installation

```bash
# Install dependencies
sudo apt update
sudo apt install -y postgresql redis-server ffmpeg

# Setup database
sudo -u postgres psql -c "CREATE DATABASE origcms;"
sudo -u postgres psql -c "CREATE USER origcms WITH PASSWORD 'password';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE origcms TO origcms;"

# Run application
go run ./cmd/server/...
```

---

## 3. Configuration

### 3.1 Configuration File

Edit `configs/bootstrap.yaml`:

```yaml
server:
  http:
    addr: 0.0.0.0:8080
  grpc:
    addr: 0.0.0.0:9000

database:
  driver: postgres
  source: "host=localhost port=5432 user=origcms dbname=origcms sslmode=disable"

redis:
  addr: localhost:6379
  password: ""
  db: 0
```

### 3.2 Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_HTTP_ADDR` | HTTP address | `0.0.0.0:8080` |
| `SERVER_GRPC_ADDR` | gRPC address | `0.0.0.0:9000` |
| `DATABASE_SOURCE` | PostgreSQL DSN | - |
| `REDIS_ADDR` | Redis address | `localhost:6379` |
| `JWT_SECRET` | JWT signing secret | - |

### 3.3 Media Upload Settings

Configure upload limits in `configs/bootstrap.yaml`:

```yaml
upload:
  max_size: 4294967296  # 4GB
  allowed_types:
    - video/mp4
    - video/webm
    - audio/mpeg
    - image/jpeg
    - image/png
```

---

## 4. User Management

### 4.1 User Roles

| Role | Permissions |
|------|-------------|
| **admin** | Full system access, user management |
| **editor** | Media review, comment moderation |
| **user** | Upload media, comment, create playlists |
| **guest** | View public content only |

### 4.2 User Management API

```bash
# List users
GET /api/v1/users?page=1&page_size=20

# Get user
GET /api/v1/users/{id}

# Create user
POST /api/v1/users
{
  "user": {
    "username": "newuser",
    "email": "newuser@example.com",
    "nickname": "New User"
  },
  "password": "securepassword"
}

# Update user
PUT /api/v1/users/{id}
{
  "user": {
    "nickname": "Updated Name"
  }
}

# Delete user
DELETE /api/v1/users/{id}

# Update user status
PATCH /api/v1/users/{id}/status
{
  "status": 1  # 1=active, 0=inactive
}
```

### 4.3 Password Reset

```bash
# Admin password reset
POST /api/v1/users/{id}/password
{
  "password": "newpassword"
}
```

---

## 5. Media Management

### 5.1 Media Workflow

```
Upload → Processing → Review → Published
                    ↓
                (Optional)
                 Rejected
```

### 5.2 Media Visibility

| Status | Description |
|--------|-------------|
| **public** | Visible to all, appears in listings |
| **unlisted** | Visible to anyone with link, not in listings |
| **private** | Only visible to uploader and admins |

### 5.3 Media Management API

```bash
# List media
GET /api/v1/medias?page=1&page_size=20&status=public

# Get media
GET /api/v1/medias/{id}

# Update media
PUT /api/v1/medias/{id}
{
  "media": {
    "title": "Updated Title",
    "description": "Updated description"
  }
}

# Delete media
DELETE /api/v1/medias/{id}

# Increment view count
POST /api/v1/medias/{id}/view
```

### 5.4 Media Encoding Status

| Status | Description |
|--------|-------------|
| **pending** | Waiting for encoding |
| **processing** | Currently encoding |
| **success** | Encoding completed |
| **failed** | Encoding failed |
| **partial** | Partially encoded (some profiles failed) |

---

## 6. Category and Tag Management

### 6.1 Categories

Categories provide hierarchical organization for media.

```bash
# List categories
GET /api/v1/categories

# Create category
POST /api/v1/categories
{
  "category": {
    "title": "Technology",
    "slug": "technology"
  }
}

# Update category
PUT /api/v1/categories/{id}
{
  "category": {
    "title": "Tech & Science"
  }
}

# Delete category
DELETE /api/v1/categories/{id}
```

### 6.2 Tags

Tags provide flat, non-hierarchical organization.

```bash
# List tags
GET /api/v1/tags?page=1&page_size=20

# Create tag
POST /api/v1/tags
{
  "tag": {
    "title": "tutorial"
  }
}

# Update tag
PUT /api/v1/tags/{id}
{
  "tag": {
    "title": "tutorials"
  }
}

# Delete tag
DELETE /api/v1/tags/{id}
```

---

## 7. Dashboard

### 7.1 Dashboard API

```bash
# Get dashboard stats
GET /api/v1/admin/stats/dashboard

Response:
{
  "total_users": 150,
  "total_medias": 450,
  "total_views": 25000,
  "total_channels": 120,
  "pending_reviews": 5
}
```

### 7.2 Media Stats

```bash
# Get media stats
GET /api/v1/admin/stats/medias?period=week

Response:
{
  "total_uploads": 45,
  "total_views": 5000,
  "daily_stats": [
    {"date": "2026-04-17", "uploads": 10, "views": 500}
  ]
}
```

### 7.3 User Stats

```bash
# Get user stats
GET /api/v1/admin/stats/users?period=month

Response:
{
  "total_users": 150,
  "new_users": 25,
  "active_users": 80
}
```

---

## 8. Transcoding

### 8.1 Encoding Profiles

Encoding profiles define output formats for video transcoding.

```bash
# List profiles
GET /api/v1/medias/encoding/profiles

# Create profile
POST /api/v1/medias/encoding/profiles
{
  "profile": {
    "name": "720p",
    "resolution": "1280x720",
    "video_codec": "h264",
    "audio_codec": "aac"
  }
}
```

### 8.2 Default Profiles

| Profile | Resolution | Bitrate | Codec |
|---------|------------|----------|-------|
| **1080p** | 1920x1080 | 8Mbps | H.264 |
| **720p** | 1280x720 | 4Mbps | H.264 |
| **480p** | 854x480 | 2Mbps | H.264 |
| **360p** | 640x360 | 1Mbps | H.264 |
| **preview** | 320x180 | 256kbps | H.264 |

### 8.3 Transcoding Status

```bash
# Get overall encoding status
GET /api/v1/medias/encoding/status

Response:
{
  "processing_count": 5,
  "pending_count": 10,
  "failed_count": 2,
  "success_count": 100
}
```

### 8.4 Retry Failed Tasks

```bash
# Retry single task
POST /api/v1/medias/encoding/retry?task_id=123

# Retry all failed for media
POST /api/v1/medias/encoding/retry-all-failed?media_id=456
```

---

## 9. Maintenance

### 9.1 Database Backup

```bash
# Backup database
pg_dump -U origcms -d origcms > backup_$(date +%Y%m%d).sql

# Restore database
psql -U origcms -d origcms < backup_20260417.sql
```

### 9.2 Media Files Backup

Media files are stored in `data/uploads/`. Backup regularly:

```bash
# Backup media files
tar -czf media_backup_$(date +%Y%m%d).tar.gz data/uploads/

# Restore media files
tar -xzf media_backup_20260417.tar.gz
```

### 9.3 Log Rotation

Configure log rotation in `/etc/logrotate.d/origcms`:

```
/var/log/origcms/*.log {
    daily
    rotate 14
    compress
    delaycompress
    missingok
    notifempty
    create 0640 root root
}
```

### 9.4 Health Check

```bash
# Check service health
curl http://localhost:8080/health

# Check database connection
go run ./cmd/server/... -check-db
```

### 9.5 Update Procedure

```bash
# Pull latest code
git pull origin main

# Regenerate code
make generate

# Run migrations
go run ./init_db.go

# Restart service
sudo systemctl restart origcms
```

---

## Appendix A: Troubleshooting

### Database Connection Issues

```
Error: dial tcp: connection refused
```

Solution: Check PostgreSQL is running and credentials are correct.

### FFmpeg Not Found

```
Error: exec: "ffmpeg": executable file not found
```

Solution: Install FFmpeg `sudo apt install ffmpeg`

### Media Upload Fails

1. Check upload size limit in configuration
2. Verify file type is allowed
3. Check disk space availability

---

*Document Version: 1.0*
*Last Updated: 2026-04-17*