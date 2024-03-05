package hezzltest

//go:generate bin/sqlc generate

//go:generate swag init -g  internal/handlers/handler.go   -o ./docs --parseDependency
