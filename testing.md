# Testing Guide

Quick reference for running tests.

## Backend Tests

```bash
make backend-test-unit          # Unit tests
make backend-test-integration   # Integration tests (requires services)
make backend-test-race          # Race detector
```

## Frontend Tests

```bash
make frontend-test-unit  # Unit tests
make frontend-test       # E2E tests
```

## Quick Health Check

```bash
curl -f http://localhost:3002/api/health
curl -f http://localhost:7700/health
```

## Release Checklist

- [ ] `make backend-test` passes
- [ ] `make frontend-test` passes
- [ ] Health endpoints respond
- [ ] Manual smoke tests pass
