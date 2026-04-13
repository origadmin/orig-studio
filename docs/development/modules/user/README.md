# User Module

> User authentication, registration, profile management.

## Module Owner

- Owner: **TBD**
- Status: **Complete**

## Documents

- [DESIGN.md](./DESIGN.md) — User system design (see project docs)
- [TESTING.md](./TESTING.md) — Acceptance criteria

## Current Implementation

| Feature | Status | Notes |
|---------|--------|-------|
| User registration | ✅ Done | `POST /api/v1/auth/signup` |
| User login (JWT) | ✅ Done | `POST /api/v1/auth/signin` |
| JWT middleware | ✅ Done | Bearer token validation |
| Password hashing | ✅ Done | bcrypt |
| User profile CRUD | ✅ Done | `GET /api/v1/auth/me` |
| RBAC (admin/user) | ✅ Done | JWT claims + middleware |
| First user = admin | ✅ Done | Auto-promotion logic |

---

*Last updated: 2026-04-13*
