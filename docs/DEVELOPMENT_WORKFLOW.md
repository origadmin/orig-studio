# Development Workflow

> How to contribute to orig-cms and manage changes.

---

## 1. Getting Started

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/m1-jwt-auth`
3. Make your changes
4. Submit a Pull Request

See [QUICK_START.md](./QUICK_START.md) for setup instructions.

## 2. Code Conventions

- **Language**: All source code (comments, logs, errors) in English
- **Commit format**: `type(scope): description`
  - `feat(svc-user): implement JWT login`
  - `fix(svc-media): fix upload timeout`
  - `docs: update README`
  - `refactor(gateway): extract middleware`
- **API prefix**: `/api/v1/`
- **Error format**: `{"code": 40001, "message": "...", "data": null}`

## 3. Adding Features

1. Check [MILESTONES.md](./MILESTONES.md) to find the relevant milestone
2. Create a feature branch from `develop`
3. Implement the feature with tests
4. Update MILESTONES.md to mark tasks complete
5. Submit PR with clear description

For major features, create a design doc in `docs/designs/` first.

## 4. Reporting Bugs

1. Open a GitHub Issue with:
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details (OS, Go version, etc.)
2. Tag with relevant milestone (M1-M5) if applicable

## 5. Plan Changes

Small changes (task reordering, adding sub-tasks): edit [MILESTONES.md](./MILESTONES.md) directly.

Large changes (cross-milestone moves, architecture decisions): document what changed, when, and why as a comment in MILESTONES.md.

## 6. Branch Strategy

```
main           ← Stable releases only
develop        ← Integration branch
feature/m*-x   ← Feature branches
```

## 7. Testing

Run tests before submitting PRs:

```bash
go vet ./...
go test ./...
go build ./...
```

---

*Last updated: 2026-04-03*
