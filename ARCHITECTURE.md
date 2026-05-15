# Architecture

## Purpose

This repository contains the STDS brand-facing lightweight backend. General
agent behavior is defined in [`AGENTS.md`](AGENTS.md), implementation rules are
defined in [`docs/.rules/coding-style.md`](docs/.rules/coding-style.md), and
test rules are defined in [`docs/.rules/testing.md`](docs/.rules/testing.md).

The service exposes a small public API surface for brand and availability data.
It is not the core STDS backend and should not own admin, tenant, lease, billing,
repair, attachment, scheduler, or accounting workflows.

## Initial Runtime Shape

Planned stack:

- Go
- Gin
- PostgreSQL access when endpoint data requires SQL
- Cloud Run
- Cloud Logging through structured stdout/stderr logs
- Secret Manager values injected as environment variables

The first version should stay close to this shape:

```text
cmd/server
internal/http
internal/application
internal/platform/database
internal/shared
```

This structure is a guide, not permission to create empty packages. Add packages
only when code exists for that responsibility.

## API Boundary

The planned API surface is only a few public brand endpoints. Endpoint scope
should be issue-driven and should stay read-only unless explicitly approved.

Public API behavior should prioritize:

- stable response shapes
- safe error messages
- low operational coupling to the core backend
- no exposure of private STDS operational data

## Data Boundary

This backend should read approved brand-facing data only. The preferred contract
is a readonly database principal querying approved views from the core STDS
database, for example brand profile, FAQ items, and property availability views.

The service should not query core operational tables directly unless a GitHub
issue explicitly documents the reason, risk, and approved boundary.

## Cloud Deployment

The intended baseline is:

```text
Browser -> Firebase Hosting rewrite -> Cloud Run brand backend -> Cloud SQL
```

Operational assumptions:

- Firebase Hosting can route brand API traffic through a path such as
  `/brand-api/**`.
- Cloud Run hosts the backend container.
- Secrets are injected from Secret Manager into environment variables.
- Database access uses a limited readonly credential.
- Cloud Logging receives structured request logs from stdout/stderr.

Private Cloud SQL connectivity, VPC egress, and exact staging/prod resource
names should be documented when the deployment issue is created.

## Observability

The baseline middleware should provide:

- request ID
- method
- path
- status
- latency
- structured JSON logs suitable for Cloud Logging

Do not log secrets, tokens, raw request bodies, database credentials, or private
operational data.

## Open Decisions

- Final endpoint names and response contracts.
- Whether the first SQL contract uses database views from day one or starts with
  temporary local/static data during early scaffolding.
- Exact Cloud Run service name, Firebase Hosting rewrite path, and staging/prod
  secret names.
