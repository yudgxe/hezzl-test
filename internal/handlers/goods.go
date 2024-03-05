package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/yudgxe/hezzl-test/internal/database/sqlc"
	"github.com/yudgxe/hezzl-test/internal/model/clickhouse"
	"github.com/yudgxe/hezzl-test/internal/tools"
	"github.com/yudgxe/hezzl-test/internal/types"
)

const goodSubj = "logs.good"

// GoodMiddleware - парсит с url project_id и id, а так же проверяет существование записи.
func (e *RouterEnv) goodMiddleware(g *gin.Context) {
	goodID, err := strconv.ParseInt(g.Query("id"), 10, 32)
	if err != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, WebError{Code: 0, Message: err.Error()})
		return
	}
	projectID, err := strconv.ParseInt(g.Query("project_id"), 10, 32)
	if err != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, WebError{Code: 0, Message: err.Error()})
		return
	}
	exist, err := e.sql().HasGood(g, sqlc.HasGoodParams{
		ID:        int32(goodID),
		ProjectID: int32(projectID),
	})
	if err != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, WebError{Code: 1, Message: err.Error()})
		return
	}
	if !exist {
		g.AbortWithStatusJSON(http.StatusNotFound, WebError{Code: 3, Message: "errors.good.notFound"})
		return
	}

	g.Set("good_id", int32(goodID))
	g.Set("project_id", int32(projectID))
	g.Next()
}

type goodCreateBody struct {
	Name string `json:"name" binding:"required" example:"name"`
}

// @Summary				Create good
// @Param               project_id query int true "Project id"
// @Param request       body goodCreateBody{} true "query params"
// @Description			Create good.
// @Produce				application/json
// @Tags				goods
// @Router              /good/create [post]
func (e *RouterEnv) goodCreate(g *gin.Context) {
	projectID, err := strconv.ParseInt(g.Query("project_id"), 10, 32)
	if err != nil {
		g.JSON(http.StatusBadRequest, WebError{Code: 0, Message: err.Error()})
		return
	}
	var body goodCreateBody
	if ok := bindAndValidate(g, &body); !ok {
		return
	}
	good, err := e.sql().CreateGood(g, sqlc.CreateGoodParams{Name: body.Name, ProjectID: int32(projectID)})
	if err != nil {
		g.JSON(http.StatusInternalServerError, WebError{Code: 1, Message: err.Error()})
		return
	}
	e.logger.Info().Interface("good", good).Msg("created good")
	g.JSON(http.StatusCreated, good)
}

type goodUpdateBody struct {
	Name        string           `json:"name" binding:"required" example:"name"`
	Description types.NullString `json:"description" example:"description" swaggertype:"string"`
}

// @Summary				Update good
// @Param               id query int true "Good id"
// @Param               project_id query int true "Project id"
// @Param request       body goodUpdateBody{} true "query params"
// @Description			Update good.
// @Produce				application/json
// @Tags				goods
// @Router              /good/update [PATCH]
func (e *RouterEnv) goodUpdate(g *gin.Context) {
	goodID := g.MustGet("good_id").(int32)
	projectID := g.MustGet("project_id").(int32)
	var body goodUpdateBody
	if ok := bindAndValidate(g, &body); !ok {
		return
	}
	good, err := e.sql().UpdateGood(g, sqlc.UpdateGoodParams{
		Name:        body.Name,
		Description: body.Description,
		ID:          int32(goodID),
		ProjectID:   int32(projectID),
	})
	if err != nil {
		g.JSON(http.StatusInternalServerError, WebError{Code: 1, Message: err.Error()})
		return
	}
	g.JSON(http.StatusOK, good)
	e.logger.Info().Interface("good", good).Msg("updated")
	if err := e.cache.SetGood(g, good, redis.KeepTTL, true); err != nil {
		e.logger.Error().Err(err).Msg("failed to updated cache")
		return
	}
	if err := e.publisher.Publish(goodSubj, clickhouse.FromGoodSQLC(good)); err != nil {
		e.logger.Error().Err(err).Msg("failed to publish")
		return
	}
}

// @Summary				Delete good
// @Param               id query int true "Good id"
// @Param               project_id query int true "Project id"
// @Description			Delete good.
// @Produce				application/json
// @Tags				goods
// @Router              /good/remove [DELETE]
func (e *RouterEnv) goodRemove(g *gin.Context) {
	goodID := g.MustGet("good_id").(int32)
	projectID := g.MustGet("project_id").(int32)
	good, err := e.sql().UpdateGoodRemoved(g, sqlc.UpdateGoodRemovedParams{
		Removed:   true,
		ID:        goodID,
		ProjectID: projectID,
	})
	if err != nil {
		g.JSON(http.StatusInternalServerError, WebError{Code: 1, Message: err.Error()})
		return
	}
	g.JSON(http.StatusOK, map[string]interface{}{
		"id":         goodID,
		"project_id": projectID,
		"removed":    true,
	})
	e.logger.Info().Interface("good", good).Msg("removed")
	if err := e.cache.SetGood(g, good, redis.KeepTTL, true); err != nil {
		e.logger.Error().Err(err).Msg("failed to update cache")
	}
	if err := e.publisher.Publish(goodSubj, clickhouse.FromGoodSQLC(good)); err != nil {
		e.logger.Error().Err(err).Msg("failed to publish")
		return
	}
}

// @Summary				List goods
// @Description			List goods.
// @Param               limit query int true "Limit" default(10)
// @Param               offset query int true "Offset" default(1)
// @Produce				application/json
// @Tags				goods
// @Router              /goods/list [GET]
func (e *RouterEnv) goodList(g *gin.Context) {
	response, np, err := e.cache.GetGoodsWithPagination(g, tools.GetPagination(g))
	if err != nil {
		log.Println(err)
	}
	if !np.HasNotFound() {
		pagination := np.Pagination()
		// TODO: хранить total/removed
		g.JSON(http.StatusOK, map[string]interface{}{
			"meta": map[string]interface{}{
				"total":   0,
				"removed": 0,
				"limit":   pagination.Limit,
				"offset":  pagination.Offset,
			},
			"goods": response,
		})
		return
	}
	// Оборачиваем в транзакцию, т.к нам важно, чтобы оба запроса работали с одним набором данных.
	tx, err := e.db.Begin(g)
	if err != nil {
		g.JSON(http.StatusInternalServerError, WebError{Code: 1, Message: err.Error()})
		return
	}
	defer tx.Rollback(context.Background())

	qtx := e.sql().WithTx(tx)
	meta, err := qtx.MetaGood(g)
	if err != nil {
		g.JSON(http.StatusInternalServerError, WebError{Code: 1, Message: err.Error()})
		return
	}

	pagination := np.Pagination()
	goods, err := qtx.ListGoods(g, sqlc.ListGoodsParams{
		Limit:  pagination.Limit,
		Offset: pagination.Offset,
	})
	if err != nil {
		g.JSON(http.StatusInternalServerError, WebError{Code: 1, Message: err.Error()})
		return
	}
	if err := tx.Commit(g); err != nil {
		g.JSON(http.StatusInternalServerError, WebError{Code: 1, Message: err.Error()})
		return
	}
	response = tools.MergeSlices(response, goods, np.MergeIndex())
	g.JSON(http.StatusOK, map[string]interface{}{
		"meta": map[string]interface{}{
			"total":   meta.Total,
			"removed": meta.Removed,
			"limit":   pagination.Limit,
			"offset":  pagination.Offset,
		},
		"goods": response,
	})

	// Обновляем кеш.
	for i, good := range goods {
		if err := e.cache.SetGoodWihtPagination(g, good, time.Second*60, false, pagination.Offset+i+1); err != nil {
			e.logger.Error().Err(err).Msg("failed to update cash")
		}
	}

}

type goodReprioritiizeBody struct {
	NewPriority int `json:"new_priority" binding:"required"`
}

// @Summary				List goods
// @Description			List goods.
// @Param               id query int true "Good id"
// @Param               project_id query int true "Project id"
// @Param request       body goodReprioritiizeBody{} true "query params"
// @Produce				application/json
// @Tags				goods
// @Router              /good/reprioritiize [PATCH]
func (e *RouterEnv) goodReprioritiize(g *gin.Context) {
	goodID := g.MustGet("good_id").(int32)
	projectID := g.MustGet("project_id").(int32)

	var body goodReprioritiizeBody
	if ok := bindAndValidate(g, &body); !ok {
		return
	}

	updated, err := e.sql().Reset(g, sqlc.ResetParams{
		ID:        goodID,
		ProjectID: projectID,
		Priority:  int32(body.NewPriority),
	})
	if err != nil {
		g.JSON(http.StatusInternalServerError, WebError{Code: 1, Message: err.Error()})
		return
	}

	g.JSON(http.StatusOK, updated)
	for _, good := range updated {
		if err := e.cache.SetGood(g, good, redis.KeepTTL, true); err != nil {
			e.logger.Error().Err(err).Msg("failed to updated cache")
		}
	}

}

// int32Query - пытается найти и распасить key в URL запросе, при ошибках пишет их в ответ и возвращает false.
func int32Query(g *gin.Context, key string) (int32, bool) {
	value, err := strconv.ParseInt(g.Query(key), 10, 32)
	if err != nil {
		g.JSON(http.StatusBadRequest, WebError{Code: 0, Message: err.Error()})
		return 0, false
	}
	return int32(value), true
}
