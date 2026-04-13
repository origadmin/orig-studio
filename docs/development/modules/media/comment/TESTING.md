# Comment System Testing

> Test plan and acceptance criteria for the comment module.

## 1. Test Structure

```
tests/unit/svc-content/comment_test.go   # biz layer unit tests
tests/integration/svc-content/comment_test.go  # repo + API integration
web/e2e/comment.spec.ts                   # E2E comment flows
```

## 2. Unit Tests (biz layer)

### 2.1 CreateComment

```go
// Success: valid text, existing media
func TestCreateComment_Success(t *testing.T) {
    uc := NewMockCommentUseCase()
    comment, err := uc.CreateComment(ctx, &Comment{
        Text:    "Great video!",
        MediaID: 1,
        UserID:  5,
    })
    require.NoError(t, err)
    assert.Equal(t, "pending", comment.Status)
    assert.NotNil(t, comment.UID)
}

// Fail: empty text
func TestCreateComment_EmptyText(t *testing.T) {
    _, err := uc.CreateComment(ctx, &Comment{
        Text:    "",
        MediaID: 1,
        UserID:  5,
    })
    require.Error(t, err)
}

// Fail: media not found
func TestCreateComment_MediaNotFound(t *testing.T) {
    _, err := uc.CreateComment(ctx, &Comment{
        Text:    "Test",
        MediaID: 9999,
        UserID:  5,
    })
    require.Error(t, err)
}
```

### 2.2 UpdateComment

```go
// Success: owner updates own comment
func TestUpdateComment_OwnerSuccess(t *testing.T) {
    comment, err := uc.UpdateComment(ctx, 100, userID=5, text="Updated text")
    require.NoError(t, err)
    assert.Equal(t, "Updated text", comment.Text)
}

// Fail: non-owner, non-admin
func TestUpdateComment_PermissionDenied(t *testing.T) {
    _, err := uc.UpdateComment(ctx, 100, userID=99, text="Hacked!")
    require.Error(t, err)
    assert.Contains(t, err.Error(), "permission denied")
}
```

### 2.3 DeleteComment

```go
// Success: owner deletes own comment
func TestDeleteComment_OwnerSuccess(t *testing.T) {
    err := uc.DeleteComment(ctx, 100, userID=5, isAdmin=false)
    require.NoError(t, err)
}

// Success: admin deletes any comment
func TestDeleteComment_AdminSuccess(t *testing.T) {
    err := uc.DeleteComment(ctx, 100, userID=99, isAdmin=true)
    require.NoError(t, err)
}

// Fail: non-owner, non-admin
func TestDeleteComment_PermissionDenied(t *testing.T) {
    err := uc.DeleteComment(ctx, 100, userID=99, isAdmin=false)
    require.Error(t, err)
}
```

### 2.4 ListMediaComments

```go
// Returns approved comments only for non-admin
func TestListMediaComments_NonAdminSeesApproved(t *testing.T) {
    comments, total, err := uc.ListMediaComments(ctx, mediaID=1, page=1, pageSize=20)
    require.NoError(t, err)
    for _, c := range comments {
        assert.Equal(t, "approved", c.Status)
    }
})
```

### 2.5 Admin Operations

```go
func TestApproveComment_Success(t *testing.T) { ... }
func TestRejectComment_Success(t *testing.T) { ... }
func TestListPending_AdminOnly(t *testing.T) { ... }
```

## 3. Integration Tests

### 3.1 API Tests

Run with: `go test ./tests/integration/svc-content/... -v`

```
POST /api/v1/medias/{id}/comments
  ✓ 201: authenticated user creates comment
  ✓ 400: empty text
  ✓ 401: unauthenticated
  ✓ 404: media not found

GET /api/v1/medias/{id}/comments
  ✓ 200: returns paginated approved comments
  ✓ 200: empty list for media with no comments
  ✓ 200: includes user info and reply_count

PUT /api/v1/comments/{id}
  ✓ 200: owner updates own comment
  ✓ 403: non-owner attempts update
  ✓ 404: comment not found

DELETE /api/v1/comments/{id}
  ✓ 204: owner deletes own comment
  ✓ 204: admin deletes any comment
  ✓ 403: non-owner, non-admin

GET /api/v1/admin/comments/pending
  ✓ 200: admin lists pending comments
  ✓ 403: non-admin gets 403

PUT /api/v1/admin/comments/{id}/approve
  ✓ 200: admin approves comment
  ✓ 403: non-admin gets 403

PUT /api/v1/admin/comments/{id}/reject
  ✓ 200: admin rejects comment
  ✓ 400: reject without reason
```

## 4. E2E Tests (Playwright)

Run with: `cd web && npx playwright test e2e/comment.spec.ts`

```typescript
describe('Comment Flow', () => {
  it('user can submit comment on media', async ({ page }) => {
    await page.goto('/watch/1');
    await page.fill('[data-testid=comment-input]', 'Great video!');
    await page.click('[data-testid=comment-submit]');
    await expect(page.locator('.comment-item')).toBeVisible();
  });

  it('pending comment shows status indicator', async ({ page }) => {
    await page.goto('/watch/1');
    await expect(page.locator('.comment-status-pending')).toBeVisible();
  });

  it('user can reply to comment', async ({ page }) => {
    await page.goto('/watch/1');
    await page.click('.comment-item >> text=Reply');
    await page.fill('[data-testid=reply-input]', 'I agree!');
    await page.click('[data-testid=reply-submit]');
    await expect(page.locator('.reply-item')).toBeVisible();
  });

  it('user can delete own comment', async ({ page }) => {
    // ...
  });

  it('admin can approve comment', async ({ page }) => {
    await page.goto('/admin/comments');
    await page.click('.pending-item >> text=Approve');
    await expect(page.locator('.approved-badge')).toBeVisible();
  });
});
```

## 5. Performance Tests

```bash
# Insert 10k comments, measure list query time
ab -n 1000 -c 10 http://localhost:9090/api/v1/medias/1/comments?page=1&page_size=20
# Target: p95 < 200ms
```

---

## 6. Acceptance Checklist

Before marking comment module as complete:

```
Backend:
  [ ] All unit tests pass (go test ./internal/svc-content/... -v)
  [ ] All integration tests pass
  [ ] go vet ./... zero warnings
  [ ] Comment status flow correct (pending -> approved/rejected)

Frontend:
  [ ] Comment section visible on /watch/:id page
  [ ] User can submit comment (authenticated)
  [ ] Comment appears immediately (optimistic) then confirms
  [ ] Reply nested display correct (max 3 levels)
  [ ] Admin comment audit page functional

i18n:
  [ ] All strings use i18n keys, no hardcoded text
```

---

*Last updated: 2026-04-13*
