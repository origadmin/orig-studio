# Content Interaction Design

> Likes, favorites, notifications, channels, playlists, tags, subscriptions.

## 1. Module Scope

All user-content interactions beyond core media management:
- Like/Dislike media
- Favorite media to playlist
- Notifications (comment, like, subscription events)
- Channels (user-created content collections)
- Playlists (user-created ordered media lists)
- Tags (media categorization keywords)
- Categories (hierarchical media classification)
- Subscriptions (user-to-channel follow)

---

## 2. Sub-module Architecture

### 2.1 Like/Dislike

```
Like (entity)
├── id        int64
├── media_id  int64  (FK)
├── user_id   int64  (FK)
├── type      string (like/dislike)
├── add_date  time
└── updated_at time

Constraints:
- UNIQUE(media_id, user_id, type) — one like/dislike per user per media
```

**Toggle Logic**:
1. Check existing like by user+media+type
2. If exists: delete (toggle off)
3. If not exists: create (toggle on)
4. Update media.like_count / dislike_count atomically

### 2.2 Favorite

```
Favorite (entity)
├── id         int64
├── media_id   int64  (FK)
├── user_id    int64  (FK)
├── playlist_id *int64 (FK, nullable — quick favorite vs playlist save)
├── add_date   time

Constraints:
- UNIQUE(media_id, user_id) — one favorite per user per media
```

### 2.3 Notification

```
Notification (entity)
├── id          int64
├── user_id     int64  (FK, recipient)
├── type        string (comment/like/subscribe)
├── actor_id    int64  (FK, who triggered)
├── media_id    *int64 (FK, nullable)
├── comment_id  *int64 (FK, nullable)
├── is_read     bool
├── add_date    time
```

**Notification triggers**:
| Event | Recipient | Type | Reference |
|-------|-----------|------|-----------|
| New comment on media | Media owner | comment | media_id, comment_id |
| Reply to comment | Comment author | comment | media_id, comment_id |
| Like on media | Media owner | like | media_id |
| New subscriber | Channel owner | subscribe | channel_id |

### 2.4 Channel

```
Channel (entity)
├── id           int64
├── uid          uuid
├── name         string
├── slug         string (unique)
├── description  string
├── user_id      int64  (FK, owner)
├── is_public    bool
├── subscriber_count int64
├── add_date     time
└── edit_date    time

ChannelMedia (M2M edge)
└── Media[] linked to channel

Subscription (entity)
├── id         int64
├── user_id    int64  (FK, subscriber)
├── channel_id int64  (FK)
├── add_date   time

Constraints:
- UNIQUE(user_id, channel_id)
```

### 2.5 Playlist

```
Playlist (entity)
├── id         int64
├── uid        uuid
├── name       string
├── slug       string
├── user_id    int64  (FK, owner)
├── is_public  bool
├── media_count int64
├── add_date   time
└── edit_date  time

PlaylistMedia (ordered M2M)
├── playlist_id  int64
├── media_id     int64
├── position     int
└── add_date     time

Constraints:
- UNIQUE(playlist_id, media_id)
- UNIQUE(playlist_id, position)
```

### 2.6 Tag

```
Tag (entity)
├── id       int64
├── name     string (unique)
├── slug     string (unique)
└── use_count int64

MediaTag (M2M edge)
└── Media[] tagged with this tag
```

### 2.7 Category

```
Category (entity)
├── id          int64
├── name        string
├── slug        string
├── parent_id   *int64 (self-ref for hierarchy)
├── media_count int64
└── add_date    time

Rules:
- Max 2-level hierarchy (category -> subcategory)
```

---

## 3. Backend Directory Structure

```
internal/svc-content/
├── biz/
│   ├── like_favorite.go       # LikeUseCase, FavoriteUseCase
│   ├── notification.go        # NotificationUseCase
│   ├── playlist_channel.go    # ChannelUseCase, PlaylistUseCase
│   ├── category_tag.go        # CategoryUseCase, TagUseCase
│   └── interfaces.go          # svc-content interfaces
├── data/
│   ├── like_favorite_repo.go
│   ├── notification_repo.go
│   ├── playlist_channel_repo.go
│   └── category_tag_repo.go
└── server/
    └── (handlers via gateway)
```

---

## 4. API Endpoints Summary

| Feature | Endpoints |
|---------|-----------|
| Like | GET/POST/DELETE `/api/v1/medias/{id}/like` |
| Favorite | GET/POST/DELETE `/api/v1/medias/{id}/favorite` |
| Notification | GET `/api/v1/notifications`, PUT `/api/v1/notifications/{id}/read` |
| Channel | CRUD `/api/v1/channels`, GET `/api/v1/channels/{id}/medias` |
| Playlist | CRUD `/api/v1/playlists`, POST/PUT/DELETE `/api/v1/playlists/{id}/medias` |
| Tag | CRUD `/api/v1/tags` |
| Category | CRUD `/api/v1/categories` |
| Subscription | POST/DELETE `/api/v1/channels/{id}/subscription` |

---

## 5. Frontend Components

```
web/src/components/portal/
├── LikeButton.tsx             # Like/dislike toggle
├── FavoriteButton.tsx        # Favorite toggle
├── NotificationBell.tsx       # Bell icon + unread count
├── NotificationList.tsx       # Dropdown/list of notifications
├── ChannelCard.tsx           # Channel preview card
├── PlaylistCard.tsx          # Playlist preview card
└── TagBadge.tsx              # Tag display chip

web/src/pages/home/
├── Channel.tsx               # Channel detail page
├── Playlist.tsx              # Playlist detail page
└── Notifications.tsx         # Full notifications page
```

---

## 6. Known Gaps

- [ ] Subscription entity exists in schema but no backend implementation
- [ ] Frontend: Like/Favorite buttons on Watch page — partially done
- [ ] Frontend: Notification bell dropdown — not started
- [ ] Frontend: Channel/Playlist detail pages — partial

---

## 7. Acceptance Criteria

### 7.1 Backend

- [ ] Like/Unlike toggles correctly, counts accurate
- [ ] Favorite/Unfavorite works
- [ ] Notifications created on comment/like/subscribe
- [ ] Notifications marked read correctly
- [ ] Channel CRUD + subscription works
- [ ] Playlist CRUD + media add/remove/reorder works
- [ ] Tag CRUD + media association works
- [ ] Category CRUD + hierarchy works

### 7.2 Frontend

- [ ] Like button shows correct state (liked/not liked)
- [ ] Like count updates optimistically
- [ ] Notification bell shows unread count
- [ ] Notification dropdown lists recent notifications
- [ ] Channel page shows channel info + media list
- [ ] Playlist page shows playlist info + ordered media
- [ ] Admin: category/tag management pages work

---

*Last updated: 2026-04-13*
