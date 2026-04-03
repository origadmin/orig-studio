# Quick Start

> New session? Read this first. This is the knowledge map that prevents duplicate work.

---

## 1. What is orig-cms?

orig-cms is an open-source media content management system built with Go (Gin + Ent ORM) and React/Next.js. It handles video/image upload, transcoding (FFmpeg), HLS streaming, and content management.

**Tech Stack**:

| Layer | Technology |
|-------|-----------|
| Backend | Go, Gin, Ent ORM, PostgreSQL |
| Frontend | React, Next.js, TypeScript, Tailwind CSS |
| Media Processing | FFmpeg, Bento4 |
| Deployment | Docker Compose |

## 2. Project Structure

```
orig-cms/
├── cmd/server/          # Monolith entry point (current active mode)
├── internal/
│   ├── data/entity/     # Ent schemas (single source of truth)
│   ├── helpers/         # FFmpeg/Bento4 helpers
│   └── svc-user/        # User service module
├── api/proto/           # gRPC/Protobuf definitions
├── configs/             # YAML configuration files
├── web/                 # Next.js frontend
├── docs/                # Documentation (you are here)
└── go.mod
```

> **Enterprise Edition**: We also offer enterprise-level services including distributed cluster deployment, dedicated support, and custom solutions. Contact us for details.

## 3. How to Run

### Prerequisites

- Go 1.22+
- Node.js 18+ (or Bun)
- PostgreSQL 15+

### Backend

```bash
# Install dependencies
go mod tidy

# Start the monolith server
go run ./cmd/server/...

# Health check
curl http://localhost:9090/healthz
```

### Frontend

```bash
cd web
npm install   # or: bun install
npm run dev   # or: bun dev
```

## 4. Key Documents

| Document | Purpose |
|----------|---------|
| [MILESTONES.md](./MILESTONES.md) | Development roadmap and task tracking |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | Architecture decisions and design |
| [MEDIA_TOOLCHAIN.md](./MEDIA_TOOLCHAIN.md) | FFmpeg/Bento4 setup guide |
| [DEVELOPMENT_WORKFLOW.md](./DEVELOPMENT_WORKFLOW.md) | Contribution and change management |
| [M2_UPLOAD_PLAN.md](./M2_UPLOAD_PLAN.md) | M2 upload feature design |
| [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) | M3 transcoding plan |

## 5. Current Status (2026-04-03)

- **M0** (Architecture prep): Done
- **M1** (Foundation): Partially done (entity unification, server entry, code quality)
- **M2** (Upload & Playback): In progress (plan exists, full flow not integrated)
- **M3-M5**: Planned

## 6. Conventions

- All source code (comments, log messages, error strings) must be in English
- REST API prefix: `/api/v1/`
- Commit format: `feat(module): description`, `fix(module): description`, `docs: description`
- Branch strategy: `main` (stable) / `develop` (integration) / `feature/m*-topic`

---

*Last updated: 2026-04-03*
