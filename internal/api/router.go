package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter(h *Handler, mw ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })

	v1 := r.Group("/api/v1", mw...)
	v1.GET("/results", h.GetResult)

	return r
}