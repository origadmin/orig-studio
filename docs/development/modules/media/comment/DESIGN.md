# Comment System Design

> User submits comment on media -> backend validation -> hierarchical storage -> audit -> frontend display.

## 1. Overview

### 1.1 Module Scope

- User submits comment/reply on a media item
- Comments support nested replies (hierarchical structure)
- Admin can audit (approve/reject/delete) any comment
- Users can only delete/edit their own comments

### 1.2 Out of Scope

- Comment reactions (like/dislike) — handled in `content` module
- Comment reports — future feature
- @ mentions — future feature

---

## 2. Architecture

### 2.1 Data Model

```
Comment (entity)
├── id          int64      (PK)
├── uid         uuid       (public identifier)
├── text        string     (max 2000 chars)
├── media_id    int64      (FK -> media)
├── user_id     int64      (FK -> user)
├── parent_id   *int64     (FK -> comment, nullable for top-level)
├── status      string     (pending|approved|rejected)
├── add_date    time
└── updated_at  time

Edges:
- comment.media    -> Media (M2O)
- comment.user    -> User (M2O)
- comment.parent   -> Comment (M2O, self-ref for replies)
- comment.replies  -> Comment[] (O2M, children)
```

### 2.2 Comment Status Flow

```
[User Submit]
      ↓
   pending
      ↓
[Admin Review] ──approve──→ approved
      │
      └──reject──→ rejected
```

- Default status on create: `pending`
- Approved comments visible to all users
- Rejected comments visible only to author + admin

### 2.3 Hierarchical Design

```
Media (id=1)
├── Comment (id=100, parent_id=null, depth=0) — "Great video!"
│   ├── Comment (id=101, parent_id=100, depth=1) — "I agree"
│   │   └── Comment (id=102, parent_id=101, depth=2) — "me too"
│   └── Comment (id=103, parent_id=100, depth=1) — "well said"
└── Comment (id=104, parent_id=null, depth=0) — "Could be better"
```

- **Max depth**: 3 levels (configurable via config)
- **Pagination**: Top-level comments are paginated; replies are loaded on-demand

---

## 3. Backend Design

### 3.1 Directory Structure

```
internal/svc-content/
├── biz/
│   └── comment.go           # CommentUseCase, Comment entity, CommentRepo interface
├── data/
│   └── comment_repo.go       # ent-based CommentRepo implementation
└── server/
    └── comment.go            # Gin handlers (thin, delegate to biz)
```

### 3.2 Repository Interface

```go
type CommentRepo interface {
    Create(ctx context.Context, c *Comment) (*Comment, error)
    Get(ctx context.Context, id int) (*Comment, error)
    Update(ctx context.Context, c *Comment) (*Comment, error)
    Delete(ctx context.Context, id int) error

    // List top-level comments for a media (approved only for non-admin)
    ListByMedia(ctx context.Context, mediaID int, page, pageSize int) ([]*Comment, int, error)
    // List all comments (admin only)
    ListAll(ctx context.Context, page, pageSize int) ([]*Comment, int, error)
    // List replies for a comment
    ListReplies(ctx context.Context, parentID int, page, pageSize int) ([]*Comment, int, error)
    // List pending comments for audit
    ListPending(ctx context.Context, page, pageSize int) ([]*Comment, int, error)
    // Count replies for a comment
    CountReplies(ctx context.Context, parentID int) (int, error)
}
```

### 3.3 UseCase Interface

```go
type CommentUseCase interface {
    CreateComment(ctx context.Context, c *Comment) (*Comment, error)
    UpdateComment(ctx context.Context, id int, userID int, text string) (*Comment, error)
    DeleteComment(ctx context.Context, id int, userID int, isAdmin bool) error
    GetComment(ctx context.Context, id int) (*Comment, error)
    ListMediaComments(ctx context.Context, mediaID int, page, pageSize int) ([]*Comment, int, error)
    ListReplies(ctx context.Context, parentID int, page, pageSize int) ([]*Comment, int, error)

    // Admin operations
    ApproveComment(ctx context.Context, id int) error
    RejectComment(ctx context.Context, id int, reason string) error
    ListPending(ctx context.Context, page, pageSize int) ([]*Comment, int, error)
}
```

### 3.4 API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/medias/{media_id}/comments` | Optional | List top-level comments |
| GET | `/api/v1/comments/{id}` | Optional | Get single comment |
| GET | `/api/v1/comments/{id}/replies` | Optional | List replies |
| POST | `/api/v1/medias/{media_id}/comments` | Required | Create comment |
| POST | `/api/v1/comments/{id}/replies` | Required | Create reply |
| PUT | `/api/v1/comments/{id}` | Required | Update own comment |
| DELETE | `/api/v1/comments/{id}` | Required | Delete own comment |
| GET | `/api/v1/admin/comments/pending` | Admin | List pending comments |
| PUT | `/api/v1/admin/comments/{id}/approve` | Admin | Approve comment |
| PUT | `/api/v1/admin/comments/{id}/reject` | Admin | Reject comment |

### 3.5 Request/Response Formats

**Create Comment Request**
```json
{
  "text": "Great video!"
}
```

**Create Comment Response**
```json
{
  "id": 100,
  "uid": "uuid-string",
  "text": "Great video!",
  "media_id": 1,
  "user_id": 5,
  "parent_id": null,
  "status": "pending",
  "add_date": "2026-04-13T18:00:00Z",
  "user": {
    "id": 5,
    "username": "john",
    "avatar_url": "https://..."
  },
  "replies": []
}
```

**List Comments Response**
```json
{
  "total": 50,
  "page": 1,
  "page_size": 20,
  "comments": [
    {
      "id": 100,
      "text": "Great video!",
      "status": "approved",
      "user": {...},
      "reply_count": 3,
      "replies": [/* first 2 replies as preview */]
    }
  ]
}
```

---

## 4. Frontend Design

### 4.1 Component Structure

```
web/src/components/portal/
├── CommentSection.tsx       # Container for all comment UI
├── CommentForm.tsx          # New comment / reply input
├── CommentItem.tsx         # Single comment display
├── CommentList.tsx         # Paginated list of comments
└── CommentReplies.tsx      # Nested reply display
```

### 4.2 Display Logic

```
Comment display:
├── Author avatar + username + add_date
├── Comment text (max 2000 chars, expandable for long text)
├── Action bar:
│   ├── Reply button (opens CommentForm inline)
│   ├── Like button + count
│   ├── Edit button (own comment only)
│   └── Delete button (own comment or admin)
└── Replies section:
    └── "View X more replies" link (if reply_count > 2)
        └── CommentItem (depth=1)
            └── CommentItem (depth=2)
                └── CommentItem (depth=3, no further nesting)
```

### 4.3 i18n Keys

All strings use i18n, no hardcoded Chinese:

```
comment.title
comment.placeholder
comment.submit
comment.reply
comment.edit
comment.delete
comment.like
comment.liked
comment.replies.count
comment.view.more
comment.status.pending
comment.status.approved
comment.status.rejected
comment.audit.approve
comment.audit.reject
comment.error.submit_failed
```

---

## 5. Acceptance Criteria

### 5.1 Functional Criteria

- [ ] User can submit a comment on any media
- [ ] Comment default status is `pending`
- [ ] Approved comments visible to all users on media page
- [ ] User can reply to any comment (max depth = 3)
- [ ] User can edit their own comment
- [ ] User can delete their own comment
- [ ] Admin can approve/reject any pending comment
- [ ] Admin can delete any comment
- [ ] Comment count displayed on media card

### 5.2 Non-Functional Criteria

- [ ] Comment text max 2000 characters (validated at API layer)
- [ ] API response time < 200ms for list queries (10k comments)
- [ ] Reply pagination loads max 3 replies inline, rest via "View more"
- [ ] Optimistic UI update on submit (show immediately, rollback on failure)

### 5.3 Test Scenarios

See [TESTING.md](./TESTING.md) for full test plan.

---

## 6. Dependencies

- **Entity**: `internal/data/entity/comment` (existing, needs schema review)
- **Media module**: CommentUseCase depends on MediaUseCase for `CheckMedia()`
- **User module**: Depends on User entity for author info
- **Config**: `comment.max_depth`, `comment.page_size`

---

## 7. Open Issues

- [ ] ent schema comment entity needs `status` field added
- [ ] CommentLike entity not yet designed (future)
- [ ] Reply notification not yet designed (depends on notification module)

---

*Last updated: 2026-04-13*
