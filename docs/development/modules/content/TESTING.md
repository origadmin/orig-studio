# Content Module Testing

> Test plan for content interaction sub-modules.

## 1. Test Structure

```
tests/unit/svc-content/
├── like_favorite_test.go
├── notification_test.go
├── playlist_channel_test.go
└── category_tag_test.go
tests/integration/content_test.go
web/e2e/
├── like.spec.ts
├── favorite.spec.ts
├── notification.spec.ts
├── channel.spec.ts
└── playlist.spec.ts
```

## 2. Unit Test Scenarios

### 2.1 Like

```go
func TestToggleLike_NewLike(t *testing.T) {
    // User has no like -> create like
    uc := NewMockCommentUseCase()
    isLiked, err := uc.ToggleLike(ctx, mediaID=1, userID=5, likeType="like")
    require.NoError(t, err)
    assert.True(t, isLiked)
    assert.Equal(t, 1, mediaRepo.GetLikeCount(1))
}

func TestToggleLike_RemoveLike(t *testing.T) {
    // User has like -> remove like
    isLiked, err := uc.ToggleLike(ctx, mediaID=1, userID=5, likeType="like")
    require.NoError(t, err)
    assert.False(t, isLiked)
    assert.Equal(t, 0, mediaRepo.GetLikeCount(1))
}

func TestToggleDislike_ReplacesLike(t *testing.T) {
    // User has "like" -> toggles "dislike" -> like removed, dislike created
    isLiked, isDisliked, err := uc.ToggleDislike(ctx, mediaID=1, userID=5)
    require.NoError(t, err)
    assert.False(t, isLiked)
    assert.True(t, isDisliked)
}
```

### 2.2 Notification

```go
func TestCreateNotification_Comment(t *testing.T) {
    // User B comments on User A's media
    // -> Notification created for User A (type=comment)
    notif, err := uc.CreateNotification(ctx, &Notification{
        UserID: userA,
        Type:   "comment",
        ActorID: userB,
        MediaID: mediaID,
    })
    require.NoError(t, err)
    assert.False(t, notif.IsRead)
}

func TestMarkAsRead(t *testing.T) {
    uc.MarkAsRead(ctx, notifID=100, userID=userA)
    notif := uc.GetNotification(ctx, 100)
    assert.True(t, notif.IsRead)
}
```

### 2.3 Channel

```go
func TestCreateChannel(t *testing.T) {
    ch, err := uc.CreateChannel(ctx, &Channel{
        Name: "My Channel",
        UserID: 5,
    })
    require.NoError(t, err)
    assert.NotEmpty(t, ch.Slug)
}

func TestSubscribeChannel(t *testing.T) {
    // User A subscribes to Channel B
    subs, err := uc.Subscribe(ctx, channelID=ch.ID, userID=5)
    require.NoError(t, err)
    assert.Equal(t, 1, subs.SubscriberCount)
}
```

## 3. E2E Test Scenarios

### 3.1 Like Flow

```typescript
it('user can like and unlike media', async ({ page }) => {
  await page.goto('/watch/1');
  await expect(page.locator('[data-testid=like-button]')).toBeVisible();
  await page.click('[data-testid=like-button]');
  await expect(page.locator('[data-testid=like-button].active')).toBeVisible();
  // Click again to unlike
  await page.click('[data-testid=like-button]');
  await expect(page.locator('[data-testid=like-button]:not(.active)')).toBeVisible();
});
```

### 3.2 Notification Flow

```typescript
it('notification bell shows unread count', async ({ page }) => {
  await page.goto('/');
  const badge = page.locator('[data-testid=notification-badge]');
  const count = await badge.textContent();
  expect(parseInt(count ?? '0')).toBeGreaterThan(0);
});

it('clicking notification marks as read', async ({ page }) => {
  await page.goto('/');
  await page.click('[data-testid=notification-bell]');
  await page.click('.notification-item >> nth=0');
  // Badge count should decrease
});
```

---

## 4. Acceptance Checklist

```
Backend:
  [ ] go test ./internal/svc-content/... -v
  [ ] Like toggle: like/dislike counts accurate
  [ ] Notifications: created on correct events
  [ ] Notifications: mark as read works
  [ ] Channel: CRUD + subscription counts correct
  [ ] Playlist: CRUD + media reorder works
  [ ] Tag/Category: CRUD + media association works

Frontend:
  [ ] Like button state correct (liked/not liked)
  [ ] Like count updates optimistically
  [ ] Notification bell shows unread count
  [ ] Notification dropdown functional
  [ ] Channel/Playlist pages render correctly
  [ ] Admin category/tag management works
```

---

*Last updated: 2026-04-13*
