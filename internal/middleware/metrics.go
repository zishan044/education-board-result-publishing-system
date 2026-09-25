package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zishan044/education-board-result-publishing-system/internal/metrics"
)

func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		metrics.Observe(c.Request.Method, route, c.Writer.Status(), time.Since(start))
	}
}