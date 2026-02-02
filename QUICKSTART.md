# Quick Start Guide

Get the application running in Docker in 3 simple steps.

## Prerequisites

- Docker and Docker Compose installed
- Ports 8484 and 5432 available

## Start the Application

```bash
# Step 1: Start all services (database + app)
make docker-up-build

# Step 2: Seed the database
make docker-seed

# Step 3: Test the API
curl http://localhost:8484/categories | jq
```

That's it! The application is now running at http://localhost:8484

## Quick Commands

```bash
# View logs
make docker-logs-app

# Stop everything
make docker-down

# Restart after code changes
make docker-up-build
```

## API Endpoints

### Get all categories
```bash
curl http://localhost:8484/categories | jq
```

### Get products with pagination
```bash
curl "http://localhost:8484/catalog?limit=5" | jq
```

### Filter by category
```bash
curl "http://localhost:8484/catalog?category=shoes" | jq
```

### Filter by price
```bash
curl "http://localhost:8484/catalog?price_less_than=10" | jq
```

### Get product details with variants
```bash
curl http://localhost:8484/catalog/PROD001 | jq
```

### Create a new category
```bash
curl -X POST http://localhost:8484/categories \
  -H "Content-Type: application/json" \
  -d '{"code":"electronics","name":"Electronics"}' | jq
```

## Development

### Local Development (without Docker)
```bash
# Start PostgreSQL
make docker-up

# Run app locally
make seed
make run
```

### Run Tests
```bash
# Unit tests
make test

# Integration tests (in Docker)
make test-integration
```

## Troubleshooting

**Port conflict?**
```bash
make docker-down
```

**Fresh start?**
```bash
make docker-down-volumes
make docker-up-build
make docker-seed
```

**See what's wrong?**
```bash
make docker-logs-app
```

For more details, see [DOCKER.md](DOCKER.md)
