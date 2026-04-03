# Architecture

> Architecture decisions and design for orig-cms.

---

## 1. Design Principles

- **Monolith first, split later**: Start as a single Go binary (`cmd/server`), extract microservices only when justified.
- **Clean separation**: Frontend (React/Next.js) communicates with backend via REST API.
- **Ent as single ORM**: All database operations use the unified Ent schema in `internal/data/entity/`.
- **Media tooling**: FFmpeg for transcoding, Bento4 for HLS/fMP4 packaging.

## 2. Backend Architecture

```
cmd/server/main.go          ← Monolith entry point
  ├── HTTP Router (Gin)      ← REST API endpoints
  ├── Middleware              ← Auth, CORS, logging
  ├── User Module            ← Registration, login, JWT
  ├── Media Module           ← Upload, listing, playback
  ├── Content Module         ← Comments, favorites, likes
  └── Media Processing       ← FFmpeg transcoding worker
```

### Data Layer

- **ORM**: Ent (schema-driven code generation)
- **Database**: PostgreSQL
- **Schema location**: `internal/data/entity/schema/`

### API Design

- All endpoints prefixed with `/api/v1/`
- JWT-based authentication
- JSON request/response format
- Standard error codes: `{"code": 40001, "message": "...", "data": null}`

## 3. Frontend Architecture

```
web/
├── pages/
│   ├── auth/        ← Login, register
│   ├── home/        ← Public-facing pages (feed, watch, search)
│   └── admin/       ← Admin dashboard
├── components/      ← Reusable UI components
├── lib/             ← API clients, utilities
└── hooks/           ← Custom React hooks
```

- **Framework**: Next.js with React
- **Styling**: Tailwind CSS
- **State management**: React Context + TanStack Query

## 4. Media Processing Pipeline

```
Upload → Store → Extract Metadata → [Transcode] → HLS Packaging → Serve
                              ↓
                         Generate Thumbnail
```

- **Transcoding**: FFmpeg with configurable profiles (360p, 720p, 1080p)
- **HLS output**: Master playlist + per-resolution m3u8 + .ts segments
- **Tool detection**: Environment variable > `tools/bin/` directory > system PATH

See [MEDIA_TOOLCHAIN.md](./MEDIA_TOOLCHAIN.md) for FFmpeg/Bento4 setup details.

## 5. Deployment

Currently supports single-node deployment via Docker Compose:

- Go backend service
- PostgreSQL database
- FFmpeg + Bento4 for media processing
- Next.js frontend (or static export)

For distributed cluster deployment and enterprise-grade infrastructure, contact us for deployment services.

---

*Last updated: 2026-04-03*
