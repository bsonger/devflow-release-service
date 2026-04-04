package api

import (
	"github.com/bsonger/devflow-service-common/httpx"
	"github.com/gin-gonic/gin"
)

type pagination struct {
	enabled  bool
	limit    int
	offset   int
	page     int
	pageSize int
}

func parsePagination(c *gin.Context) (pagination, error) {
	p, err := httpx.ParsePagination(c)
	if err != nil {
		return pagination{}, err
	}
	return pagination{
		enabled:  p.Enabled,
		limit:    p.Limit,
		offset:   p.Offset,
		page:     p.Page,
		pageSize: p.PageSize,
	}, nil
}

func paginateSlice[T any](items []T, p pagination) []T {
	return httpx.PaginateSlice(items, httpx.Pagination{
		Enabled:  p.enabled,
		Limit:    p.limit,
		Offset:   p.offset,
		Page:     p.page,
		PageSize: p.pageSize,
	})
}

func setPaginationHeaders(c *gin.Context, total int, p pagination) {
	httpx.SetPaginationHeaders(c, total, httpx.Pagination{
		Enabled:  p.enabled,
		Limit:    p.limit,
		Offset:   p.offset,
		Page:     p.page,
		PageSize: p.pageSize,
	})
}

func includeDeleted(c *gin.Context) bool {
	return httpx.IncludeDeleted(c)
}
