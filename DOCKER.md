# Running with Docker

This guide explains how to run the application using Docker and Docker Compose.

## Quick Start

### 1. Start the Application

```bash
# Build and start all services (database + app)
make docker-up-build

# Or just start (uses cached images)
make docker-up
```

The application will be available at http://localhost:8484

### 2. Seed the Database

```bash
make docker-seed
```

### 3. Verify It's Running

```bash
# Check service status
make docker-ps

# View logs
make docker-logs

# View only app logs
make docker-logs-app
```

### 4. Test the API

```bash
# Get all categories
curl http://localhost:8484/categories | jq

# Get catalog with pagination
curl "http://localhost:8484/catalog?limit=5" | jq

# Get product by code
curl http://localhost:8484/catalog/PROD001 | jq
```

### 5. Stop the Application

```bash
# Stop all services
make docker-down

# Stop and remove volumes (clean slate)
make docker-down-volumes
```

## Available Make Commands

| Command | Description |
|---------|-------------|
| `make docker-build` | Build the Docker image |
| `make docker-up` | Start services in detached mode |
| `make docker-up-build` | Build and start services |
| `make docker-down` | Stop all services |
| `make docker-down-volumes` | Stop services and remove volumes |
| `make docker-logs` | View logs from all services |
| `make docker-logs-app` | View logs from app only |
| `make docker-seed` | Run database migrations |
| `make docker-restart` | Restart the app service |
| `make docker-ps` | Show running containers |

## Architecture

The Docker setup consists of two services:

### 1. PostgreSQL Database
- Image: `postgres:17.5`
- Port: `5432:5432`
- Health checks enabled
- Persistent volume for data

### 2. Application
- Built from `Dockerfile`
- Port: `8484:8484`
- Auto-restarts unless stopped
- Waits for database to be healthy

### Network

Both services run on a shared bridge network (`app-network`) allowing them to communicate using service names.

## Environment Variables

The application uses these environment variables (configured in `docker-compose.yml`):

| Variable | Value | Description |
|----------|-------|-------------|
| `POSTGRES_HOST` | `postgres` | Database hostname (service name) |
| `POSTGRES_PORT` | `5432` | Database port |
| `POSTGRES_USER` | `postgres` | Database user |
| `POSTGRES_PASSWORD` | `password` | Database password |
| `POSTGRES_DB` | `challenge` | Database name |
| `HTTP_PORT` | `8484` | HTTP server port |

## Development Workflow

### Making Code Changes

1. Edit your code
2. Rebuild and restart:
   ```bash
   make docker-up-build
   ```

### Debugging

View real-time logs:
```bash
make docker-logs-app
```

Execute commands inside the container:
```bash
docker compose exec app sh
```

### Database Access

Connect to PostgreSQL:
```bash
docker compose exec postgres psql -U postgres -d challenge
```

## Dockerfile Details

The application uses a **multi-stage build**:

### Stage 1: Builder
- Base: `golang:1.24.3-alpine`
- Installs dependencies
- Builds the application binary
- Optimized for build speed

### Stage 2: Runtime
- Base: `alpine:latest`
- Minimal image size (~15MB)
- Only contains binary and SQL migrations
- No build tools or source code

## Troubleshooting

### Port Already in Use

If port 8484 or 5432 is already in use:

```bash
# Stop any running services
make docker-down

# Or change ports in docker-compose.yml
```

### Database Connection Issues

Ensure the database is healthy:
```bash
docker compose ps
```

Check database logs:
```bash
docker compose logs postgres
```

### Application Crashes

View application logs:
```bash
make docker-logs-app
```

Restart the app:
```bash
make docker-restart
```

### Clean Restart

Remove everything and start fresh:
```bash
make docker-down-volumes
make docker-up-build
make docker-seed
```

## Production Considerations

For production deployments:

1. **Change default passwords** in `docker-compose.yml`
2. **Use environment files** instead of hardcoded values
3. **Enable TLS/HTTPS** with a reverse proxy (nginx, traefik)
4. **Use secrets management** for sensitive data
5. **Set resource limits** in docker-compose.yml:
   ```yaml
   deploy:
     resources:
       limits:
         cpus: '0.5'
         memory: 512M
   ```
6. **Use health checks** for monitoring
7. **Configure logging** drivers for centralized logs

## Integration Tests

To run integration tests in Docker:

```bash
make test-integration
```

This uses a separate `docker-compose.test.yml` configuration with isolated services.

## CI/CD Integration

Example GitHub Actions workflow:

```yaml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run integration tests
        run: make test-integration
      - name: Build application
        run: make docker-build
```
