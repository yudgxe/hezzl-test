make ./bin/goose 
docker compose up -d 
./bin/goose -dir migrations/postgres postgres "postgresql://postgres:postgres@"$(docker compose port postgres 5432)"/postgres?sslmode=disable" up 
