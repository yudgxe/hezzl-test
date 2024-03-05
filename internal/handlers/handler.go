package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"github.com/nats-io/nats.go"
	"github.com/yudgxe/hezzl-test/internal/database/sqlc"
	"github.com/yudgxe/hezzl-test/internal/tools"

	"github.com/rs/zerolog"

	"gopkg.in/validator.v2"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           hezzl API
// @version         1.0
// @description     hezzl-test api backend server
// @host            localhost:8084
// @BasePath        /api/v1
// @contact.name    API Support
// @contact.url     http://www.swagger.io/support
// @contact.email   support@swagger.io
// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html
func Urls(db DBTX, cache Cache, logger *zerolog.Logger, nats *nats.EncodedConn, r *gin.Engine) *gin.Engine {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	env := &RouterEnv{
		db:        db,
		cache:     cache,
		logger:    logger,
		publisher: nats,
	}

	v1 := r.Group("/api/v1")
	{
		gg := v1.Group("/good")
		{
			gg.POST("/create", env.goodCreate)
			gg.PATCH("/update", env.goodMiddleware, env.goodUpdate)
			gg.DELETE("/remove", env.goodMiddleware, env.goodRemove)
			gg.PATCH("/reprioritiize", env.goodMiddleware, env.goodReprioritiize)
		}

		v1.GET("/goods/list", env.goodList)
	}
	return r
}

// DBTX - интерфейс для создания Queries и транзакций.
type DBTX interface {
	sqlc.DBTX

	Begin(ctx context.Context) (pgx.Tx, error)
}

var _ Cache = (*tools.Cache)(nil)

// Cache - интерфейс для добавлени/получение кеша.
type Cache interface {
	ScanKey(ctx context.Context, match string) (string, bool, error)

	SetGood(ctx context.Context, good sqlc.Good, expiration time.Duration, ifexist bool) error
	SetGoodWihtPagination(ctx context.Context, good sqlc.Good, expiration time.Duration, ifexist bool, pagination int) error

	GetGoodsWithPagination(ctx context.Context, pagination tools.Pagination) ([]sqlc.Good, *tools.GetGoodsWithPaginationReponse, error)
}

type RouterEnv struct {
	db        DBTX
	cache     Cache
	publisher *nats.EncodedConn
	logger    *zerolog.Logger
}

func (e *RouterEnv) sql() *sqlc.Queries {
	return sqlc.New(e.db)
}

// bindAndValidate - биндит и валидирует body, при ошибках пишет их в ответ и возвращает false.
func bindAndValidate(g *gin.Context, body interface{}) bool {
	if err := g.ShouldBindJSON(&body); err != nil {
		g.JSON(http.StatusBadRequest, WebError{Code: 0, Message: err.Error()})
		return false
	}
	if err := validator.Validate(&body); err != nil {
		g.JSON(http.StatusBadRequest, WebError{Code: 0, Message: err.Error()})
		return false
	}
	return true
}
