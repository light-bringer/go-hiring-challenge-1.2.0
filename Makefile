tidy ::
	@go mod tidy && go mod vendor

seed ::
	@go run cmd/seed/main.go

run ::
	@go run cmd/server/main.go

test ::
	@go test -v -count=1 -race ./... -coverprofile=coverage.out -covermode=atomic

test-integration ::
	@docker compose -f docker-compose.yml up -d
	@sleep 2
	@go test -v -count=1 ./test/integration/...
	@$(MAKE) test-integration-cleanup

test-integration-cleanup ::
	@docker compose -f docker-compose.yml down -v

# Docker commands
docker-build ::
	@docker compose build

docker-up ::
	@docker compose up -d

docker-up-build ::
	@docker compose up -d --build

docker-down ::
	@docker compose down

docker-down-volumes ::
	@docker compose down -v

docker-logs ::
	@docker compose logs -f

docker-logs-app ::
	@docker compose logs -f app

docker-seed ::
	@docker compose exec app ./seed

docker-restart ::
	@docker compose restart app

docker-ps ::
	@docker compose ps
