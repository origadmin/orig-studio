# Common Infrastructure Design

> Shared designs for database, configuration, error handling, logging, and i18n.

## 1. Database

### 1.1 ORM

**Tool**: Ent (code-generated ORM)
**Schema location**: `internal/data/entity/schema/`
**Generated code**: `internal/data/entity/{entity}/`

**Schema naming conventions**:
- Table name: singular, snake_case (e.g., `files_comment`, `media_media`)
- Fields: snake_case in DB, camelCase in Go struct
- ID field: `id int64` (auto-increment)
- UUID field: `uid uuid.UUID` (public identifier for API exposure)
- Timestamps: `add_date time.Time`, `edit_date time.Time`
- Soft delete: use `deleted_at *time.Time` field

### 1.2 Repository Pattern

Every data entity has a corresponding `Repo` interface in `biz/`:

```go
// Example: Comment
type CommentRepo interface {
    Create(ctx context.Context, c *Comment) (*Comment, error)
    Get(ctx context.Context, id int) (*Comment, error)
    Update(ctx context.Context, c *Comment) (*Comment, error)
    Delete(ctx context.Context, id int) error
    ListByMedia(ctx context.Context, mediaID int, page, pageSize int) ([]*Comment, int, error)
    ListAll(ctx context.Context, page, pageSize int) ([]*Comment, int, error)
}
```

Repository implementations live in `data/` and only interact with Ent generated code.

**Rule**: Handlers and UseCases MUST NOT import Ent generated code directly. They operate through Repo interfaces.

### 1.3 Transaction Handling

```go
// Use ent.Tx() for multi-step operations
func (r *commentRepo) CreateWithNotification(ctx context.Context, c *Comment) error {
    return r.client.Tx(ctx, func(tx *ent.Tx) error {
        commentRepo := NewCommentRepo(tx)
        _, err := commentRepo.Create(ctx, c)
        if err != nil {
            return err
        }
        // Create notification
        notifRepo := NewNotificationRepo(tx)
        _, err = notifRepo.Create(ctx, &Notification{...})
        return err
    })
}
```

---

## 2. Configuration

### 2.1 Config Loading

**Tool**: `origadmin/runtime` bootstrap
**Config file**: `configs/bootstrap.yaml`

```yaml
server:
  host: "0.0.0.0"
  port: 9090

database:
  dsn: "postgres://user:pass@localhost:5432/origcms?sslmode=disable"

media:
  upload_dir: "./data/uploads"
  max_file_size: 2147483648  # 2GB
  allowed_types:
    - "video/mp4"
    - "video/webm"
    - "image/jpeg"

ffmpeg:
  path: "ffmpeg"
  ffprobe_path: "ffprobe"

jwt:
  secret: "${JWT_SECRET}"
  expiry_hours: 72

comment:
  max_depth: 3
  page_size: 20
```

### 2.2 Accessing Config

```go
import "origadmin/runtime/config"

cfg := config.Get()
log.Infof("Server port: %d", cfg.Server.Port)
```

---

## 3. Error Handling

### 3.1 Error Response Format

All API errors return JSON in this format:

```json
{
  "code": 40001,
  "message": "Error description in English",
  "data": null
}
```

**Error codes**:
| Range | Category |
|-------|----------|
| 40000-49999 | Client errors (4xx) |
| 50000-59999 | Server errors (5xx) |

**Standard codes**:
| Code | Meaning |
|------|---------|
| 40001 | Bad request (validation failed) |
| 40101 | Unauthorized (not logged in) |
| 40301 | Forbidden (no permission) |
| 40401 | Not found |
| 40901 | Conflict (e.g., duplicate) |
| 50001 | Internal server error |

### 3.2 Error Creation

```go
import "github.com/origadmin/toolkits/errors"

func (uc *CommentUseCase) CreateComment(ctx context.Context, c *Comment) (*Comment, error) {
    if c.Text == "" {
        return nil, errors.BadRequest("comment text is required")
    }
    // ...
    if err != nil {
        return nil, errors.Internal("failed to create comment: %v", err)
    }
}
```

### 3.3 Handler Error Mapping

```go
// internal/server/error.go
func mapError(err error) (int, *ErrorResponse) {
    switch {
    case errors.IsBadRequest(err):
        return 400, &ErrorResponse{Code: 40001, Message: err.Error()}
    case errors.IsUnauthorized(err):
        return 401, &ErrorResponse{Code: 40101, Message: err.Error()}
    case errors.IsForbidden(err):
        return 403, &ErrorResponse{Code: 40301, Message: err.Error()}
    case errors.IsNotFound(err):
        return 404, &ErrorResponse{Code: 40401, Message: err.Error()}
    default:
        return 500, &ErrorResponse{Code: 50001, Message: "internal error"}
    }
}
```

---

## 4. Logging

**Tool**: `github.com/go-kratos/kratos/v2/log`

```go
import "github.com/go-kratos/kratos/v2/log"

// In UseCase
log := log.NewHelper(log.With(logger, "module", "comment.biz"))
log.Infof("Comment created: id=%d, media_id=%d", comment.ID, comment.MediaID)
log.Warnf("Media not found: media_id=%d", c.MediaID)
log.Errorf("Failed to create comment: %v", err)
```

**Log levels**: Debug < Info < Warn < Error
**Structured fields**: Always include `module` field (e.g., `module=comment.biz`)

---

## 5. i18n (Internationalization)

### 5.1 Backend i18n

Use `github.com/origadmin/toolkits/i18n` or simple key-based approach:

```go
// For user-facing error messages, return English keys
return nil, errors.NotFound("comment not found")

// For multi-language support, use i18n package
msg := i18n.Translate(ctx, "comment.error.not_found", locale)
return nil, errors.NotFound(msg)
```

### 5.2 Frontend i18n

**Tool**: `react-i18next`
**Config**: `web/src/i18n/index.ts`
**Namespaces**: `auth`, `common`, `home`, `watch`, `admin`, `comment`

```typescript
// Usage
import { useTranslation } from 'react-i18next';

function CommentForm() {
  const { t } = useTranslation('comment');
  return <button>{t('comment.submit')}</button>;
}
```

**Key naming convention**: `{page}.{component}.{action}` or `{page}.{status}`

```
comment.submit
comment.placeholder
comment.error.submit_failed
watch.subscribe
watch.like
```

**Missing keys**: If a key resolves to the raw key (e.g., `watch.subscribe` shown instead of "Subscribe"), check:
1. Namespace is loaded: `i18n.loadNamespaces(['comment', 'watch'])`
2. Translation file exists: `web/public/locales/{lang}/{ns}.json`
3. Key exists in translation file

---

## 6. Code Style Rules

### 6.1 No Chinese Comments

**Enforced rule**: All code comments MUST be in English.

```go
// GOOD
// CreateComment creates a new comment for the given media.
func (uc *CommentUseCase) CreateComment(...) { ... }

// BAD
// 创建评论
func (uc *CommentUseCase) CreateComment(...) { ... }
```

### 6.2 Error Messages

Error messages shown to users (in API response `message` field) should be in English, concise, and actionable.

```go
// GOOD
return nil, errors.BadRequest("comment text cannot be empty")
return nil, errors.NotFound("media not found")

// BAD
return nil, errors.BadRequest("评论内容不能为空")
return nil, errors.NotFound("媒体不存在")
```

### 6.3 Log Messages

Log messages should be in English and include relevant context fields.

```go
// GOOD
log.Infof("Comment created: id=%d media_id=%d user_id=%d", c.ID, c.MediaID, c.UserID)

// BAD
log.Infof("创建评论成功")
```

---

## 7. Testing Standards

### 7.1 Test File Location

Unit tests: same directory as the file under test
Integration tests: `tests/integration/`
E2E tests: `web/e2e/`

### 7.2 Test Naming

```go
// Format: Test{Method}_{Scenario}
func TestCreateComment_Success(t *testing.T) { ... }
func TestCreateComment_EmptyText(t *testing.T) { ... }
func TestCreateComment_MediaNotFound(t *testing.T) { ... }
func TestCreateComment_PermissionDenied(t *testing.T) { ... }
```

### 7.3 Test Dependencies

Mock repositories in unit tests. Use interfaces:

```go
type MockCommentRepo struct {
    comments map[int]*Comment
}

func (m *MockCommentRepo) Create(ctx context.Context, c *Comment) (*Comment, error) {
    c.ID = len(m.comments) + 1
    m.comments[c.ID] = c
    return c, nil
}

// In test
uc := NewCommentUseCase(&MockCommentRepo{}, mediaUC)
```

---

## 8. CI Requirements

All PRs must pass:

```bash
go vet ./...           # No warnings
go test ./...          # All tests pass
go build ./...         # Compiles without error
cd web && bun run lint # ESLint + Prettier (no errors)
```

---

*Last updated: 2026-04-13*
