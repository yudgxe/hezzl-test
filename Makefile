
.PHONY:run
run:
	docker compose up -d
	go run cmd/api/main.go -dbdsn="postgresql://postgres:postgres@localhost:$(docker compose port postgres 5432 | awk -F: '{ print $$2 }')/postgres"

bin/goose:
	env GOBIN=$$(pwd)/bin go install -v github.com/pressly/goose/v3/cmd/goose@latest

bin/sqlc:
	env GOBIN=$$(pwd)/bin go install -v github.com/sqlc-dev/sqlc/cmd/sqlc@latest

bin/swag:
	env GOBIN=$$(pwd)/bin go install -v github.com/swaggo/swag/cmd/swag@latest

.PHONY:generate
generate: bin/sqlc bin/swag
	go generate ./...

.PHONY: migrate
migrate:
	docker compose up -d
	go run migrations/clickhouse/main.go db migrate
	./bin/goose -dir migrations/postgres postgres "postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable" up