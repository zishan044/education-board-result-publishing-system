package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zishan044/education-board-result-publishing-system/internal/cache"
	"github.com/zishan044/education-board-result-publishing-system/internal/result"
)

type StatsGetter interface {
	Stats(ctx context.Context, exam string, year int16) (*result.Stats, error)
}

type StatsHandler struct {
	store   StatsGetter
	cache   *cache.Cache
	ttl     time.Duration
	timeout time.Duration
}

func NewStatsHandler(s StatsGetter, c *cache.Cache, ttl, timeout time.Duration) *StatsHandler {
	return &StatsHandler{store: s, cache: c, ttl: ttl, timeout: timeout}
}

func (h *StatsHandler) GetStats(c *gin.Context) {
	exam := c.DefaultQuery("exam", "SSC")
	year, err := strconv.ParseInt(c.DefaultQuery("year", "2026"), 10, 16)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()

	cacheKey := fmt.Sprintf("%s:%d", exam, year)
	var cached result.Stats
	if hit, err := h.cache.GetStats(ctx, cacheKey, &cached); err == nil && hit {
		c.Header("Cache-Control", "public, max-age=60")
		c.JSON(http.StatusOK, cached)
		return
	}

	st, err := h.store.Stats(ctx, exam, int16(year))
	if err != nil {
		slog.Error("get stats", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if err := h.cache.SetStats(ctx, cacheKey, st, h.ttl); err != nil {
		slog.Error("cache stats", "err", err)
	}

	c.Header("Cache-Control", "public, max-age=60")
	c.JSON(http.StatusOK, st)
}

