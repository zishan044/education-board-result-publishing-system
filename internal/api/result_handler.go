package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zishan044/education-board-result-publishing-system/internal/result"
	"github.com/zishan044/education-board-result-publishing-system/internal/store"
)

type ResultGetter interface {
	Get(ctx context.Context, k result.Key) (*result.Result, error)
}

type Handler struct {
	results ResultGetter
	timeout time.Duration
}

func NewHandler(r ResultGetter, timeout time.Duration) *Handler {
	return &Handler{results: r, timeout: timeout}
}

func (h *Handler) GetResult(c *gin.Context) {
	key, err := parseKey(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()

	res, err := h.results.Get(ctx, key)
	switch {
	case errors.Is(err, store.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "result not found"})
	case err != nil:
		slog.Error("get result", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	default:
		c.Header("Cache-Control", "public, max-age=3600")
		c.JSON(http.StatusOK, res)
	}
}

func parseKey(c *gin.Context) (result.Key, error) {
	exam, board := c.Query("exam"), c.Query("board")
	if exam == "" || board == "" {
		return result.Key{}, errors.New("exam and board are required")
	}
	year, err := strconv.ParseInt(c.Query("year"), 10, 16)
	if err != nil {
		return result.Key{}, fmt.Errorf("invalid year")
	}
	roll, err := strconv.ParseInt(c.Query("roll"), 10, 32)
	if err != nil {
		return result.Key{}, fmt.Errorf("invalid roll")
	}
	reg, err := strconv.ParseInt(c.Query("registration"), 10, 64)
	if err != nil {
		return result.Key{}, fmt.Errorf("invalid registration")
	}
	return result.Key{
		Exam: exam, ExamYear: int16(year), Board: board,
		Roll: int32(roll), Registration: reg,
	}, nil
}

