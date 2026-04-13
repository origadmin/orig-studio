# Coding Standards

> Enforced coding conventions for orig-cms. These rules are NOT suggestions — they are requirements.

---

## 1. Language

**All code, comments, logs, and error messages MUST be in English.**

This includes:
- Source code and comments
- Commit messages
- API error messages
- Log output
- Documentation
- i18n keys (keys can be in English, values translated via i18n)

---

## 2. Code Organization

### 2.1 Module Structure

Each module follows this structure:

```
internal/svc-{name}/
├── biz/           # Business logic, use cases, entities
│   ├── biz.go     # Module initialization, wire providers
│   ├── {feature}.go
│   └── {feature}_test.go
├── data/          # Data access, repository implementations
│   ├── data.go
│   └── {feature}_repo.go
├── dto/           # Data transfer objects
│   └── {feature}.go
├── server/        # Gin HTTP handlers (thin)
│   └── {feature}.go
└── service/       # HTTP <-> biz conversion
    └── {feature}.go
```

### 2.2 File Naming

- Go files: `snake_case.go` (e.g., `comment_repo.go`, `media_handler.go`)
- Test files: same name with `_test.go` suffix
- Frontend: `PascalCase.tsx` for components, `camelCase.ts` for utilities

---

## 3. Go Conventions

### 3.1 Imports

Standard library first, then third-party, then internal:

```go
package comment

import (
    "context"
    "errors"
    "time"

    "github.com/go-kratos/kratos/v2/log"
    "github.com/google/uuid"

    "origadmin/application/origcms/internal/data/entity"
)
```

### 3.2 Naming

| Type | Convention | Example |
|------|-----------|---------|
| Package | lowercase, no underscores | `biz`, `data`, `helpers` |
| Struct | PascalCase | `CommentUseCase`, `MediaRepo` |
| Interface | PascalCase with `er` suffix or noun | `CommentRepo`, `UseCase` |
| Function | PascalCase | `CreateComment`, `ListMedias` |
| Variable | camelCase | `commentList`, `mediaID` |
| Constant | PascalCase | `MaxFileSize`, `DefaultPageSize` |
| Private field | camelCase | `repo`, `logger` |
| Error variable | `Err{Description}` | `ErrNotFound`, `ErrPermissionDenied` |

### 3.3 Error Handling

```go
// GOOD: Descriptive error with context
if comment.Text == "" {
    return nil, errors.New("comment text is required")
}

// GOOD: Wrapped error with context
created, err := r.repo.Create(ctx, comment)
if err != nil {
    return nil, fmt.Errorf("create comment: %w", err)
}

// BAD: Bare error
if err != nil {
    return nil, err
}
```

### 3.4 Context Usage

Always pass `context.Context` as the first parameter:

```go
func (uc *CommentUseCase) CreateComment(ctx context.Context, c *Comment) (*Comment, error)
```

Never store `ctx` in structs. Never use `context.Background()` in request handlers.

### 3.5 Logging

```go
// Include module field
log := log.NewHelper(log.With(logger, "module", "comment.biz"))

// Log with structured fields
log.Infof("Comment created: id=%d media_id=%d", comment.ID, comment.MediaID)
log.Warnf("Media not found: media_id=%d", mediaID)
log.Errorf("Failed to create comment: %v", err)
```

---

## 4. API Design

### 4.1 REST Endpoints

- Prefix: `/api/v1/`
- Method: REST verbs (`GET`, `POST`, `PUT`, `DELETE`)
- Collection: `/medias`, single item: `/medias/{id}`
- Sub-resource: `/medias/{id}/comments`

### 4.2 Request/Response

```json
// Request (application/json)
{ "title": "My Video", "description": "Description" }

// Success response
{ "data": { "id": 1, "title": "My Video" }, "code": 0, "message": "ok" }

// Error response
{ "code": 40001, "message": "title is required", "data": null }
```

### 4.3 Pagination

```json
{
  "data": { "items": [...], "total": 100, "page": 1, "page_size": 20 },
  "code": 0,
  "message": "ok"
}
```

Query params: `?page=1&page_size=20`

---

## 5. Database

### 5.1 Ent Schema

- Use `uid` (UUID) for public-facing IDs
- Use `id` (int64) for internal FK references
- Timestamps: `add_date`, `edit_date`
- Soft delete: `deleted_at *time.Time`

### 5.2 Repository Pattern

Always define interfaces in `biz/` and implement in `data/`:

```go
// biz/comment.go
type CommentRepo interface {
    Create(ctx context.Context, c *Comment) (*Comment, error)
    Get(ctx context.Context, id int) (*Comment, error)
    // ...
}

// data/comment_repo.go
type commentRepo struct {
    client *ent.Client
}

func NewCommentRepo(client *ent.Client) CommentRepo {
    return &commentRepo{client: client}
}
```

---

## 6. Testing

### 6.1 Test-Driven

Write tests BEFORE implementation for new features.

### 6.2 Unit Test Structure

```go
func TestCreateComment_Success(t *testing.T) {
    uc := NewMockUseCase()
    comment, err := uc.CreateComment(ctx, &Comment{Text: "test"})
    require.NoError(t, err)
    assert.Equal(t, "test", comment.Text)
}

func TestCreateComment_EmptyText(t *testing.T) {
    _, err := uc.CreateComment(ctx, &Comment{Text: ""})
    require.Error(t, err)
}
```

### 6.3 Coverage Targets

| Layer | Target |
|-------|--------|
| biz (business logic) | ≥ 80% |
| service (HTTP conversion) | ≥ 60% |
| data (repository) | ≥ 40% |

---

## 7. Git Workflow

### 7.1 Branch Naming

```
feature/{module}-{task}   # feature/comment-api
bugfix/{issue-id}         # bugfix/123-comment-crash
hotfix/{description}      # hotfix/urgent-fix
```

### 7.2 Commit Messages

```
type(module): description

feat(comment): add nested reply support
fix(media): correct thumbnail path on delete
docs: update comment module design doc
test(comment): add unit tests for create/update/delete
refactor(svc-content): extract notification logic
```

### 7.3 Pull Requests

- Title: Clear description of what changed
- Description: Link to relevant issue/design doc
- CI must pass before merge
- At least 1 reviewer approval

---

## 8. Frontend

### 8.1 Component Structure

- Components: `PascalCase.tsx` in `components/`
- Hooks: `use{Noun}.ts` in `hooks/`
- Utilities: `camelCase.ts` in `lib/`
- API clients: `{noun}.ts` in `lib/api/`

### 8.2 i18n Keys

Format: `{page}.{component}.{action}` or `{page}.{status}`

```
comment.submit
comment.placeholder
comment.error.text_required
watch.like
watch.dislike
admin.media.delete
```

### 8.3 No Hardcoded Text

All user-visible strings must use i18n:

```tsx
// GOOD
<Button>{t('comment.submit')}</Button>

// BAD
<Button>Submit</Button>
<Button>评论</Button>
```

---

## 9. Documentation

### 9.1 Code Documentation

```go
// CreateComment creates a new comment on a media item.
// The comment is created with status "pending" and requires admin approval
// before becoming visible to other users.
func (uc *CommentUseCase) CreateComment(ctx context.Context, c *Comment) (*Comment, error)
```

### 9.2 README Files

Each module must have a `README.md` with:
- Module owner
- Current status
- Link to design doc
- Feature checklist

### 9.3 Design Documents

Each feature module must have a `DESIGN.md` containing:
- Overview and scope
- Data model (with ER diagram text)
- API endpoints
- Acceptance criteria
- Open issues

---

## 10. Review Checklist

Before submitting PR, verify:

```
Code:
  [ ] No Chinese comments or strings
  [ ] Error messages in English
  [ ] Log messages in English with structured fields
  [ ] Tests written for new features
  [ ] go vet ./... passes
  [ ] go test ./... passes
  [ ] go build ./... passes

API:
  [ ] Endpoints follow REST conventions
  [ ] Error codes follow standard format
  [ ] Pagination implemented where needed

Frontend:
  [ ] All strings use i18n (no hardcoded text)
  [ ] ESLint passes
  [ ] TypeScript compiles without errors

Documentation:
  [ ] README updated if module structure changed
  [ ] Design doc updated if API changed
```

---

*Last updated: 2026-04-13*
