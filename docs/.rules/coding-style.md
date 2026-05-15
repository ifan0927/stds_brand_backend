# Coding Style

## Scope

This document defines lightweight implementation rules for the STDS brand backend.
General agent behavior is defined in [`AGENTS.md`](../../AGENTS.md).

This service is intentionally small. Prefer clear code, boring dependencies, and
direct flow over framework-heavy architecture.

## Project Shape

- Use Go and Gin unless the issue explicitly changes the stack.
- Keep handlers thin: parse request, call the service/repository path, and write
  the response.
- Keep business rules, response assembly, and database access out of handlers.
- Add interfaces, helpers, or packages only when repetition or testability
  justifies them.

## Error Handling

- Use one public API error shape across the service.
- Keep public error messages safe, short, and stable.
- Map expected failures clearly: validation, not found, unavailable, and
  internal errors should not collapse into one generic path.
- Wrap unexpected internal errors with useful context, but do not expose
  sensitive details in responses.

## Logging And Middleware

- Use structured logging suitable for Cloud Run logs.
- Every request should have a request ID.
- Request logs should include method, path, status, latency, and request ID.
- Do not log secrets, database credentials, bearer tokens, raw request bodies,
  or private STDS operational data.

## Database Boundary

- Treat this backend as read-only unless an issue explicitly approves a write.
- Prefer approved brand-facing views or equivalent read-only contracts.
- Do not query core operational tables directly unless the issue documents the
  boundary and reason.

## Go Style

- Follow idiomatic Go and the surrounding local style.
- Keep names explicit and comments minimal.
- Use comments only when the code's intent is not obvious.
- Run `gofmt` on Go files before reporting completion.
