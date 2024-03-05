testbin=gotest
ifeq (, $(shell which gotest))
testbin=go test
endif


.PHONY:build
build:
	docker compose -f build/docker-compose.yml build

.PHONY:publish
publish:
	docker compose -f build/docker-compose.yml push

.PHONY:run
run:
	docker compose up -d
	go run cmd/api/main.go -dbdsn="postgresql://postgres:postgres@localhost:$(docker compose port postgres 5432 | awk -F: '{ print $$2 }')/postgres"

.PHONY:test
test:
	$(testbin) ./...

.PHONY:docs/index.html
docs/index.html:
	swag init -g internal/handlers/handler.go  -o ./docs --parseDependency

bin/goose:
	env GOBIN=$$(pwd)/bin go install -v github.com/pressly/goose/v3/cmd/goose@latest

bin/sqlc:
	env GOBIN=$$(pwd)/bin go install -v github.com/sqlc-dev/sqlc/cmd/sqlc@latest

bin/swag:
	env GOBIN=$$(pwd)/bin go install -v github.com/swaggo/swag/cmd/swag@latest

.PHONY:generate
generate: bin/sqlc bin/swag
	go generate ./...
