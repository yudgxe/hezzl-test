package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yudgxe/hezzl-test/internal/handlers"
	"github.com/yudgxe/hezzl-test/internal/model/clickhouse"
	"github.com/yudgxe/hezzl-test/internal/tools"

	_ "github.com/yudgxe/hezzl-test/docs"
)

var (
	host string
	port int

	redisHost     string
	redisPort     int
	redisPassword string

	batchSize int

	dbdsn string
)

func init() {
	dbdsn = os.Getenv("DBDSN")
	flag.StringVar(&dbdsn, "dbdsn", dbdsn, "DSN строка для соединения с БД")

	flag.StringVar(&host, "host", "localhost", "хост для сервера")
	flag.IntVar(&port, "port", 8084, "порт для сервера")

	flag.StringVar(&redisHost, "redis-host", "localhost", "хост для редиса")
	flag.IntVar(&redisPort, "redis-port", 6379, "порт для редиса")
	flag.StringVar(&redisPassword, "redis-password", "redis", "пароль от редиса")

	flag.IntVar(&batchSize, "batch-size", 10, "размера батча логов для оправки в clickhouse")
	flag.Parse()
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	conn, err := ch.Open(&ch.Options{
		Addr: []string{"localhost:9000"},
		Auth: ch.Auth{
			Database: "logs",
			Username: "default",
			Password: "",
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to open clickhouse")
		return
	}

	if err := conn.Ping(context.Background()); err != nil {
		log.Error().Err(err).Msg("failed to ping clickhouse")
		return
	}

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Printf("%s\n", err)
		return
	}
	defer nc.Close()

	ec, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Error().Err(err).Msg("failed to get nats json encoder")
		return
	}

	tools.NewWorker[clickhouse.Good](ec, tools.NewGoodSender(conn), batchSize).Start("logs.good")

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisHost, redisPort),
		Password: redisPassword,
		DB:       0,
	})

	// health checkse
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Error().Err(err).Msg("failed to ping redis")
	}

	cfg, err := pgxpool.ParseConfig(dbdsn)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse dbdsn")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	pool, err := pgxpool.ConnectConfig(ctx, cfg)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to db")
	}
	defer pool.Close()

	r := gin.New()
	r.Use(gin.Recovery())

	client.FlushAll(context.TODO())

	cache := tools.NewCache(client)
	if err := handlers.Urls(pool, cache, &log.Logger, ec, r).Run(fmt.Sprintf("%s:%d", host, port)); err != nil {
		log.Error().Err(err).Str("host", host).Int("port", port).Msg("failed to start server")
	}
}
