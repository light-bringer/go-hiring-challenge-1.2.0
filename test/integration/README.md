# Integration Tests

This directory contains integration tests that verify the entire application stack end-to-end.

## Overview

The integration tests:
- Start a real PostgreSQL database
- Run database migrations
- Start the HTTP server
- Make actual HTTP requests to test endpoints
- Verify responses and database state

## Running Tests

### Option 1: With Docker Compose (Recommended)

Run the tests in an isolated containerized environment:

```bash
make test-integration
```

This will:
1. Build a test container with all dependencies
2. Start a PostgreSQL container
3. Run migrations
4. Execute all integration tests
5. Clean up containers automatically

To manually clean up after tests:

```bash
make test-integration-cleanup
```

### Option 2: Locally

Prerequisites:
- PostgreSQL running on localhost:5432
- Database seeded with `make seed`

```bash
go test -v -count=1 ./test/integration/...
```

## Test Coverage

The integration tests cover:

### Catalog Endpoints
- **GET /catalog**
  - Pagination (`?offset=N&limit=M`)
  - Category filtering (`?category=CODE`)
  - Price filtering (`?price_less_than=AMOUNT`)
  - Combined filters
  - Invalid parameters handling

- **GET /catalog/{code}**
  - Successful product retrieval with variants
  - Price inheritance (variants without prices inherit from product)
  - Not found handling

### Categories Endpoints
- **GET /categories**
  - List all categories

- **POST /categories**
  - Create new category
  - Missing fields validation
  - Duplicate code detection

### End-to-End Scenarios
- Product with category relationship verification
- Data consistency across endpoints

## Test Structure

```
test/integration/
├── README.md           # This file
└── api_test.go         # All integration tests
```

## Environment Variables

The tests use these environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `POSTGRES_HOST` | Database host | `localhost` |
| `POSTGRES_PORT` | Database port | `5432` |
| `POSTGRES_USER` | Database user | `postgres` |
| `POSTGRES_PASSWORD` | Database password | `password` |
| `POSTGRES_DB` | Database name | `challenge` |
| `HTTP_PORT` | Server port for tests | `8485` |

## CI/CD Integration

Add to your CI pipeline:

```yaml
# GitHub Actions example
- name: Run integration tests
  run: make test-integration
```

## Troubleshooting

**Port conflicts**: The tests use port 8485 for the HTTP server and 5433 for the database (mapped from container's 5432). If you see port conflicts, check for running services.

**Database connection issues**: Ensure PostgreSQL is healthy before running tests. The Docker Compose setup includes health checks.

**Test failures**: Run `make test-integration-cleanup` to ensure a clean state before retrying.
