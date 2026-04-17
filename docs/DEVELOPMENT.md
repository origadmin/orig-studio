# Developer Documentation

## Table of Contents

- [1. Welcome](#1-welcome)
- [2. Development Environment](#2-development-environment)
- [3. Project Structure](#3-project-structure)
- [4. Code Standards](#4-code-standards)
- [5. API Development](#5-api-development)
- [6. Testing](#6-testing)
- [7. Building and Deployment](#7-building-and-deployment)

---

## 1. Welcome

This document is for OrigCMS developers and contains information about setting up the development environment, code standards, and best practices.

### Prerequisites

- Go 1.21+
- Node.js 18+ (for frontend)
- PostgreSQL 15+
- FFmpeg (for video transcoding)
- Git

---

## 2. Development Environment

### 2.1 Clone and Setup

```bash
# Clone the repository
git clone https://github.com/origadmin/orig-cms.git
cd orig-cms

# Install development tools
make init

# Generate code (protobuf, ent, wire)
make generate

# Run database migrations
go run ./init_db.go
```

### 2.2 Configuration

Create a configuration file in `configs/` or use environment variables:

```yaml
# configs/bootstrap.yaml
server:
  http:
    addr: 0.0.0.0:8080
  grpc:
    addr: 0.0.0.0:9000

database:
  driver: postgres
  source: "host=localhost port=5432 user=postgres dbname=origcms sslmode=disable"

redis:
  addr: localhost:6379
```

### 2.3 Running the Server

```bash
# Development mode (monolithic)
go run ./cmd/server/...

# Frontend development
cd web && npm install && npm run dev
```

### 2.4 Docker Development

```bash
# Build Docker image
make build-docker

# Run with Docker Compose
docker compose up
```

---

## 3. Project Structure

### 3.1 Backend Structure (Go)

```
internal/
├── svc-{module}/           # Service module
│   ├── biz/               # Business logic layer
│   │   ├── {module}.go   # Use case implementation
│   │   └── *_test.go     # Unit tests
│   ├── data/             # Data access layer
│   │   ├── {module}_repo.go  # Repository implementation
│   │   └── provider.go   # Dependency provider
│   ├── dto/              # Data transfer objects
│   ├── service/          # gRPC/HTTP handlers
│   │   └── {module}_service.go
│   └── server/           # Server registration
│
├── data/                  # Shared data layer
│   ├── entity/           # Ent ORM schemas
│   │   ├── schema/       # Schema definitions
│   │   └── generated/    # Generated code
│   └── enums/            # Enumerations
│
├── server/                # HTTP/gRPC server
└── helpers/              # Shared utilities
```

### 3.2 Frontend Structure (React)

```
web/src/
├── api/                   # API client layer
│   ├── index.ts          # Barrel export
│   ├── media.ts          # Media API
│   └── comment.ts        # Comment API
│
├── lib/                  # Utility functions
│   ├── request.ts        # Axios wrapper
│   ├── auth.ts           # Auth utilities
│   └── format.ts         # Formatters
│
├── components/            # UI components
│   ├── ui/               # shadcn/ui base components
│   ├── auth/             # Authentication components
│   ├── media/            # Media components
│   └── admin/            # Admin components
│
├── pages/                # Page containers
│   ├── admin/            # Admin pages
│   ├── auth/             # Auth pages
│   └── home/             # Home pages
│
├── hooks/                # React hooks
├── routes/               # TanStack Router config
├── types/                # TypeScript types
└── i18n/                # Internationalization
```

---

## 4. Code Standards

### 4.1 Go Code Standards

#### Language

- **English only** in code, comments, logs, and error messages
- **Context first**: All functions should accept `context.Context` as the first parameter
- **Wrap errors**: Use `fmt.Errorf("...: %w", err)` for error wrapping

```go
// ✅ Good
func (uc *UserUseCase) GetUser(ctx context.Context, id string) (*User, error) {
    user, err := uc.repo.Get(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get user %s: %w", id, err)
    }
    return user, nil
}

// ❌ Bad
func (uc *UserUseCase) GetUser(id string) (*User, error) {
    user, err := uc.repo.Get(context.Background(), id)
    if err != nil {
        return nil, err  // Error not wrapped
    }
    return user, nil
}
```

#### Logging

Use structured logging with the `log` package:

```go
log.Context(ctx).Infof("User %s created successfully", userID)
log.Context(ctx).WithValues("user_id", userID).Infof("Login successful")
```

#### Import Organization

Imports should be organized in three groups:

1. Standard library
2. Third-party packages
3. Internal packages

```go
import (
    "context"
    "fmt"

    "github.com/go-kratos/kratos/v2/log"
    "github.com/origadmin/runtime/errors"

    "origadmin/application/origcms/internal/data/entity"
)
```

### 4.2 Frontend Code Standards

#### Language

- **English only** in code, comments, and user-visible text
- **i18n for all user-visible text**
- **No Chinese characters in source code**

#### TypeScript

- **No `any` types** — Use proper typing
- **Use `@/` for cross-module imports**
- **Use `./` for same-directory imports only**
- **No `../` parent directory imports**

```typescript
// ✅ Good
import { mediaApi, Media } from '@/api/media'
import { formatDate } from '@/lib/format'
import { Button } from '@/components/ui/button'

// ❌ Bad
import { something } from '../utils'        // No ../ allowed
import { other } from '../../lib/api'     // No ../ allowed
import { data } from './types'            // No any type
```

#### Component Structure

```typescript
// components/media/MediaCard.tsx
import { type MediaCardProps } from './types'

export function MediaCard({ media, onClick }: MediaCardProps) {
    return (
        <div onClick={onClick}>
            {/* Component content */}
        </div>
    )
}
```

---

## 5. API Development

### 5.1 Protobuf Definition

API definitions are in `api/proto/v1/`:

```protobuf
// api/proto/v1/user/user_service.proto
service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {
    option (google.api.http) = {
      get: "/api/v1/users/{id}"
    };
  }

  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {
    option (google.api.http) = {
      post: "/api/v1/users"
      body: "*"
    };
  }
}

message GetUserRequest {
  string id = 1;
}

message GetUserResponse {
  User user = 1;
}
```

### 5.2 Code Generation

After modifying proto files, regenerate code:

```bash
# Generate all code
make generate

# Generate specific
make gen-proto    # Generate protobuf
make gen-ent     # Generate ent ORM
make gen-wire    # Generate wire dependencies
```

### 5.3 API Response Format

Standard JSON response format:

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

Error response:

```json
{
  "code": 404,
  "message": "user not found",
  "data": null
}
```

---

## 6. Testing

### 6.1 Backend Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./internal/svc-user/... -v

# Run with race detector
go test -race ./...
```

### 6.2 Frontend Testing

```bash
# Install dependencies
cd web && npm install

# Run tests
npm test

# Run with coverage
npm run test:coverage
```

### 6.3 Integration Testing

```bash
# Run integration tests
go test ./tests/integration/...

# Run e2e tests
go test ./tests/e2e/...
```

### 6.4 Test Structure

```
internal/svc-user/
├── biz/
│   ├── user.go
│   └── user_test.go      # Unit tests
└── data/
    ├── user_repo.go
    └── user_repo_test.go # Integration tests
```

---

## 7. Building and Deployment

### 7.1 Build Commands

```bash
# Build binary
make build

# Build Docker image
make build-docker

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o server ./cmd/server
```

### 7.2 Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_HTTP_ADDR` | HTTP server address | `0.0.0.0:8080` |
| `SERVER_GRPC_ADDR` | gRPC server address | `0.0.0.0:9000` |
| `DATABASE_SOURCE` | PostgreSQL connection string | - |
| `REDIS_ADDR` | Redis address | `localhost:6379` |
| `JWT_SECRET` | JWT signing secret | - |

### 7.3 Deployment Checklist

- [ ] All tests passing
- [ ] Code generated (`make generate`)
- [ ] Configuration file created
- [ ] Database migrated
- [ ] Environment variables set
- [ ] FFmpeg installed (for transcoding)

---

## 8. Common Tasks

### 8.1 Adding a New Entity

1. Define schema in `internal/data/entity/schema/`
2. Run `make gen-ent`
3. Add repository in `internal/svc-{module}/data/`
4. Add use case in `internal/svc-{module}/biz/`
5. Add service handler in `internal/svc-{module}/service/`
6. Add proto definitions in `api/proto/v1/`
7. Run `make generate`
8. Add tests

### 8.2 Adding a New API Endpoint

1. Add proto definition in `api/proto/v1/`
2. Run `make gen-proto`
3. Implement service handler
4. Register route in `internal/server/routes.go`
5. Add tests

### 8.3 Debugging

```bash
# Enable debug logging
LOG_LEVEL=debug go run ./cmd/server/...

# Connect with Delve
dlv debug ./cmd/server
```

---

*Document Version: 1.0*
*Last Updated: 2026-04-17*