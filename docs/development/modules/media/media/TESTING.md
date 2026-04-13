# Media Base System Testing

> Acceptance criteria and test plan for media base module.

## 1. Test Structure

```
tests/unit/svc-media/
├── biz/
│   ├── media_test.go        # MediaUseCase unit tests
│   ├── upload_test.go       # Upload logic tests
│   └── transcode_worker_test.go
tests/integration/
├── media_test.go            # Full upload + encode flow
web/e2e/
├── upload.spec.ts           # Upload flow E2E
└── watch.spec.ts            # Watch page E2E
```

## 2. Backend Unit Tests

### 2.1 MediaUseCase

```go
// ListMedias: pagination, filtering
func TestListMedias_Pagination(t *testing.T) { ... }
func TestListMedias_FilterByType(t *testing.T) { ... }
func TestListMedias_FilterByStatus(t *testing.T) { ... }

// UpdateMedia: ownership check
func TestUpdateMedia_OwnerSuccess(t *testing.T) { ... }
func TestUpdateMedia_NonOwnerForbidden(t *testing.T) { ... }
func TestUpdateMedia_AdminCanUpdate(t *testing.T) { ... }

// DeleteMedia: cascade files
func TestDeleteMedia_CascadeFiles(t *testing.T) { ... }
```

### 2.2 Upload

```go
func TestChunkAssembly_Complete(t *testing.T) {
    // Upload 10 chunks, complete, verify file integrity
}
func TestChunkAssembly_Resume(t *testing.T) {
    // Upload 3 chunks, pause, resume from chunk 4
}
func TestChunkAssembly_MissingChunk(t *testing.T) {
    // Complete with missing chunk 5 -> error
}
```

## 3. E2E Tests

### 3.1 Upload Flow

```typescript
it('user can upload video and see it in list', async ({ page }) => {
  await page.goto('/me/upload');
  const fileInput = page.locator('input[type=file]');
  await fileInput.setInputFiles('test-data/sample.mp4');
  await page.click('[data-testid=upload-submit]');
  // Wait for encoding to complete (poll status)
  await expect(page.locator('.upload-success')).toBeVisible({ timeout: 60000 });
  // Navigate to home and verify media appears
  await page.goto('/');
  await expect(page.locator('.media-card')).toContainText('sample.mp4');
});
```

### 3.2 Watch Page

```typescript
it('watch page displays i18n text correctly', async ({ page }) => {
  await page.goto('/watch/1');
  // Verify no raw i18n keys visible
  await expect(page.locator('text=watch.subscribe')).toHaveCount(0);
  await expect(page.locator('text=watch.like')).toHaveCount(0);
});

it('owner sees edit/delete buttons', async ({ page }) => {
  await page.goto('/watch/1');
  // If current user is owner
  await expect(page.locator('[data-testid=edit-media]')).toBeVisible();
  await expect(page.locator('[data-testid=delete-media]')).toBeVisible();
});
```

---

## 3. Acceptance Checklist

```
Upload:
  [ ] go test ./internal/svc-media/... -v (all pass)
  [ ] Chunked upload resume works
  [ ] File type validation enforced

Playback:
  [ ] HLS playback works
  [ ] Quality switcher works
  [ ] MP4 fallback for failed media works

Frontend:
  [ ] Watch page: no raw i18n keys
  [ ] Watch page: owner actions visible
  [ ] Upload: progress bar accurate
  [ ] Upload: error handling (size limit, type error)
```

---

*Last updated: 2026-04-13*
