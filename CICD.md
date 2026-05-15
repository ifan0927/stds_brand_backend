# CI/CD Strategy

## Goal

Provide the minimum delivery guardrails for the STDS brand backend while keeping
the project lightweight. General agent behavior is defined in [`AGENTS.md`](AGENTS.md).

This service is not the core backend. It should not inherit heavy checks for
legacy migration, scheduler workflows, generated OpenAPI sync, or broad E2E
acceptance unless those responsibilities are explicitly added later.

## Branch Flow

Use the same basic environment flow as the core STDS backend:

```text
feature branch -> dev -> staging -> prod
```

- Feature work starts from `dev`.
- Pull requests merge into `dev`.
- `staging` is the deployment validation branch.
- `prod` represents production release state.

## PR Checks

Recommended minimum PR checks:

- `go test ./...`
- `go build ./...`
- formatting check, if the repository later adds one
- basic static checks, if they stay fast and low-maintenance

The PR gate should stay fast enough to run on every change.

## Staging Deployment

Staging should prove that the service starts and can reach its configured
dependencies.

Minimum staging validation:

- deploy the service image
- confirm `/health`
- confirm structured request logs appear in Cloud Logging
- confirm database connectivity if the deployed version uses SQL
- run a small read-only smoke request for each public endpoint

## Production Deployment

Production deployment should stay manual or explicitly approved until the
staging path is stable.

Minimum production validation:

- confirm staging passed
- deploy the same reviewed code path
- confirm `/health`
- run read-only smoke checks
- inspect recent Cloud Run logs for startup errors and repeated 5xx responses

## Non-Goals

- No Kubernetes.
- No Jenkins.
- No canary or blue-green deployment in the first version.
- No exhaustive E2E suite for three public endpoints.
- No production write smoke tests unless the service later grows write behavior.
