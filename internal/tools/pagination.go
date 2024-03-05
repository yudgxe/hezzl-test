package tools

import "github.com/gin-gonic/gin"

const (
	paginationDefaultLimit  = 10
	paginationDefaultOffset = 1
)

type Pagination struct {
	Limit  int `json:"limit" form:"limit"  example:"10"`
	Offset int `json:"offset" form:"offset" example:"1"`
}

// GetPage - биндит Pagination, при ошибке возвращает значания по умолчанию.
func GetPagination(g *gin.Context) Pagination {
	p := Pagination{}
	if g.BindQuery(&p) != nil {
		p.Limit = paginationDefaultLimit
		p.Offset = paginationDefaultOffset
	}
	return p
}
