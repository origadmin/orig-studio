# Frontend Test Directory Structure

> **Project**: orig-cms/web
> **Tech Stack**: Bun + Rsbuild + Jest + React + Playwright
> **Reference**: Backend `tests/README.md`

---

## Directory Structure

```
web/tests/
├── README.md                        ← This file
│
├── unit/                            ← Unit tests (isolated, no external deps)
│   ├── hooks/                       ← Hook unit tests
│   │   ├── useAuth.test.ts
│   │   ├── useMedia.test.ts
│   │   ├── usePagination.test.ts
│   │   └── useUpload.test.ts
│   ├── lib/                         ← Library/utility unit tests
│   │   ├── api.test.ts              ← API client tests
│   │   ├── format.test.ts           ← Format utility tests
│   │   └── auth.test.ts             ← Auth utility tests
│   └── components/                  ← Component unit tests
│       ├── ui/                      ← shadcn/ui component tests
│       ├── common/                  ← Shared component tests
│       └── ...
│
├── integration/                     ← Integration tests (component + store + API mock)
│   ├── auth.test.tsx                ← Auth flow integration
│   ├── media.test.tsx               ← Media CRUD integration
│   └── ...
│
├── e2e/                             ← E2E tests (Playwright, browser-level)
│   ├── auth.spec.ts                 ← Auth flow E2E
│   └── watch.spec.ts                ← Watch flow E2E
│
├── features/                        ← Feature-organized tests (mirrors backend)
│   └── Fxxx-{name}/
│       ├── TEST_COVERAGE.md          ← Required: test coverage declaration
│       ├── TEST_CASES.md             ← Required: specific test cases
│       ├── unit/                     ← Feature unit tests
│       │   └── *.test.{ts,tsx}
│       ├── integration/              ← Feature integration tests
│       │   └── *.test.{ts,tsx}
│       └── e2e/                      ← Feature E2E tests
│           └── *.spec.ts
│
├── bugs/                            ← Bug-organized tests (mirrors backend)
│   └── Bxxx-{name}/
│       ├── TEST_CASE.md              ← Required: reproduction test case
│       └── regression_*.test.{ts,tsx}← Regression tests
│
└── mocks/                           ← Shared mock configurations
    ├── handlers.ts                   ← MSW request handlers
    ├── server.ts                     ← MSW server setup
    └── fixtures/                     ← Test data fixtures
        ├── users.ts
        ├── media.ts
        └── ...
```

---

## Naming Rules

### Feature tests (features/)

| File | Description | Required |
|------|------------|----------|
| `TEST_COVERAGE.md` | Test coverage declaration | YES |
| `TEST_CASES.md` | Specific test cases | YES |
| `*.test.{ts,tsx}` | Test code | YES |

**Directory naming**: `F{xxx}-{short-name}` (e.g. `F014-unified-pagination`)

### Bug tests (bugs/)

| File | Description | Required |
|------|------------|----------|
| `TEST_CASE.md` | Reproduction and regression test case | YES |
| `regression_*.test.{ts,tsx}` | Regression test code | YES |

**Directory naming**: `B{xxx}-{short-name}` (e.g. `B001-media-status-bug`)

---

## Test Type Definitions

### Unit Tests

- **Framework**: Jest + ts-jest
- **Component testing**: @testing-library/react + @testing-library/jest-dom
- **Location**: `tests/unit/` or `tests/features/Fxxx/unit/`
- **Run**: `bun run test`
- **Coverage target**: 70% (minimum 60%)

**What to test**:
- Hook logic (state transitions, return values, edge cases)
- Utility functions (format, parse, validate)
- Component rendering (correct output, props handling)
- Component interactions (click, input, form submit)

### Integration Tests

- **Framework**: Jest + MSW + @testing-library/react
- **Location**: `tests/integration/` or `tests/features/Fxxx/integration/`
- **Run**: `bun run test`

**What to test**:
- Component + TanStack Query integration
- Component + Auth context integration
- Component + Router integration
- API mock → Component rendering flow

### E2E Tests

- **Framework**: Playwright
- **Location**: `tests/e2e/` or `tests/features/Fxxx/e2e/`
- **Run**: `bun run test:e2e`

**What to test**:
- Full user flows (login → browse → watch → interact)
- Cross-page navigation
- Auth-protected route access
- Responsive layout verification

---

## Required Test Categories (Feature)

> Minimum requirements for each Feature

| # | Test Category | Coverage Status | Test File | Notes |
|---|--------------|----------------|-----------|-------|
| 1 | **Component rendering** | ☐ / ☑ | `component.test.tsx` | Renders correctly with props |
| 2 | **User interaction** | ☐ / ☑ | `interaction.test.tsx` | Click/input/submit behavior |
| 3 | **Hook logic** | ☐ / ☑ | `hook.test.ts` | State transitions, return values |
| 4 | **API integration (MSW)** | ☐ / ☑ | `integration.test.tsx` | Mock API → component flow |
| 5 | **Edge/boundary values** | ☐ / ☑ | `boundary.test.tsx` | Empty/null/long/special chars |
| 6 | **Error handling** (optional) | ☐ / ☑ | `error.test.tsx` | 4xx/5xx/network error |
| 7 | **Responsive layout** (optional) | ☐ / ☑ | `responsive.test.tsx` | Breakpoint behavior |
| 8 | **Accessibility** (optional) | ☐ / ☑ | `a11y.test.tsx` | ARIA, keyboard nav |

---

## Relationship with Backend Test Structure

| Backend | Frontend | Mapping |
|---------|----------|---------|
| `tests/unit/` | `web/tests/unit/` | 1:1 |
| `tests/integration/` | `web/tests/integration/` | 1:1 |
| `tests/e2e/` | `web/tests/e2e/` | 1:1 |
| `tests/features/Fxxx/` | `web/tests/features/Fxxx/` | 1:1 |
| `tests/bugs/Bxxx/` | `web/tests/bugs/Bxxx/` | 1:1 |
| `tests/api/` | `web/tests/integration/` | API tests → integration |
| N/A | `web/tests/unit/components/` | Frontend-specific |
| N/A | `web/tests/mocks/` | Frontend-specific |

---

## Template Files

Test document templates are in `projects/team-flow/team/templates/`:

| Template | Usage |
|----------|-------|
| `projects/team-flow/team/templates/feature-test-template.md` | Feature test coverage declaration |
| `projects/team-flow/team/templates/bug-test-template.md` | Bug reproduction test case |
| `projects/team-flow/team/templates/frontend-feature-test-template.md` | Frontend Feature test (new) |
| `projects/team-flow/team/templates/frontend-bug-test-template.md` | Frontend Bug test (new) |

---

## Trigger Conditions

| Change Type | Create Directory | Use Template |
|------------|-----------------|-------------|
| Feature development | `tests/features/Fxxx/` | `frontend-feature-test-template.md` |
| Bug fix | `tests/bugs/Bxxx/` | `frontend-bug-test-template.md` |

---

## Jest Configuration

Tests in `web/tests/` are configured via `jest.config.ts` at `web/` root:

```typescript
testMatch: [
  '**/*.test.ts',
  '**/*.test.tsx',
  '<rootDir>/tests/**/*.test.ts',
  '<rootDir>/tests/**/*.test.tsx',
],
```

**Note**: Existing test files in `src/` are grandfathered. New tests MUST be placed in `tests/`.

---

## Prohibitions

- ❌ Creating scattered test files in `src/` (organize under `tests/`)
- ❌ Test files without corresponding documentation (code + doc must coexist)
- ❌ Skipping TEST_COVERAGE.md / TEST_CASE.md before writing test code
- ❌ Using Vitest imports (use Jest: `import { describe, it, expect } from '@jest/globals'`)
- ❌ Using Vue/Vite patterns (use React + Rsbuild patterns)
