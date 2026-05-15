# Testing Rules

## Scope

This document defines lightweight test rules for the STDS brand backend. General
agent behavior is defined in [`AGENTS.md`](../../AGENTS.md), and implementation
boundaries are defined in [`coding-style.md`](coding-style.md).

The goal is confidence for a small public backend, not broad coverage metrics or
core-backend-level release machinery.

## Source Of Truth

Before writing or changing tests, read:

1. The GitHub issue scope and acceptance criteria.
2. [`ARCHITECTURE.md`](../../ARCHITECTURE.md) and relevant project docs.
3. Existing tests for the package, when they exist.
4. Implementation details only after the expected behavior is clear.

If docs and implementation conflict, report the mismatch instead of encoding the
current behavior as correct.

## Test Selection

- Handler or router changes should have HTTP boundary tests for status codes,
  response shape, and important parameter handling.
- Service logic should have focused unit tests.
- Repository code should have contract tests for SQL args, scan behavior, and
  not-found/error mapping.
- Middleware should test request ID propagation, structured log fields where
  practical, and error response behavior.
- Bug fixes should include a reproducing test whenever practical.

## Test Style

- Use Go's standard `testing` package unless the project already chose another
  pattern.
- Keep assertions explicit and behavior-focused.
- Do not add a new assertion library for a narrow change.
- Prefer small fakes over real external providers.
- Do not call real GCP, Firebase, email, or other external services in normal
  tests.

## PR Gate

Every PR should state:

- what changed
- which tests were added or updated
- which test command was run
- why tests were not added, if the change is docs-only or otherwise low risk

Expected minimum checks after the project has Go code:

```bash
go test ./...
go build ./...
```
