# Project Analysis

> Technical overview and design decisions for orig-cms.

---

## 1. Overview

orig-cms is a media content management system that handles video/image upload, transcoding, streaming, and content management. This document captures the key design decisions and their rationale.

## 2. Key Design Decisions

### 2.1 Monolith-First Strategy

**Decision**: Start with a single binary (`cmd/server`) before splitting into microservices.

**Rationale**: premature microservice splitting adds complexity without benefit when the project is still finding its product-market fit. The monolith can be split later when clear boundaries emerge from actual usage patterns.

### 2.2 Gin over Kratos

**Decision**: Use Gin as the HTTP framework for the monolith.

**Rationale**: Simpler setup, lower barrier for contributors, sufficient for single-service architecture.

### 2.3 Ent ORM

**Decision**: Use Ent as the sole ORM, with a single schema package.

**Rationale**: Schema-driven code generation ensures type safety. Having one schema location (`internal/data/entity/`) avoids field drift between packages.

### 2.4 FFmpeg + Bento4

**Decision**: Use FFmpeg for transcoding and Bento4 for HLS/fMP4 packaging.

**Rationale**: Industry-standard tools with broad codec support. Bento4 produces fMP4-based HLS which is more efficient than MPEG-TS.

See [MEDIA_TOOLCHAIN.md](./MEDIA_TOOLCHAIN.md) for tool configuration.

### 2.5 svc-portal Removal

**Decision**: Aggregate API logic merged into the API gateway layer.

**Rationale**: A separate portal microservice added deployment overhead without clear benefit in monolith mode.

## 3. Frontend Architecture

- **Framework**: Next.js with React
- **Styling**: Tailwind CSS
- **UI Components**: shadcn/ui
- **Route Structure**: Three layers:
  - `pages/auth/` — Authentication pages
  - `pages/home/` — User-facing pages
  - `pages/admin/` — Admin dashboard

## 4. Database Schema

16 Ent schemas defined in `internal/data/entity/schema/`:

`media`, `user`, `category`, `tag`, `comment`, `playlist`, `encode_profile`, `encoding_task`, `upload_session`, `channel`, `favorite`, `like`, `notification`, `media_category`, `media_tag`, `media_playlist`

## 5. Known Limitations

- No object storage support yet (local filesystem only)
- No full-text search (database LIKE queries)
- No real-time notifications (polling only)
- Single-node deployment only

See [MILESTONES.md](./MILESTONES.md) for planned improvements.

---

*Last updated: 2026-04-03*
