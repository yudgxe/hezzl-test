package handlers

import (
	"github.com/gin-gonic/gin"
)

// WebError - ошибка хендлеров.
type WebError struct {
	Code    int
	Message string
	Details interface{}
}

func handleError(g *gin.Context, err error) {
	// TODO: ставим коды, мессаджи и детали в зависимости от типа err.
}
