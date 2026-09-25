package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zishan044/education-board-result-publishing-system/internal/cache"
	"github.com/zishan044/education-board-result-publishing-system/internal/pdfgen"
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

type PDFHandler struct {
	results ResultGetter
	pdfRoot string
	timeout time.Duration
}

func NewPDFHandler(r ResultGetter, pdfRoot string, timeout time.Duration) *PDFHandler {
	return &PDFHandler{results: r, pdfRoot: pdfRoot, timeout: timeout}
}

func (h *PDFHandler) GeneratePDF(c *gin.Context) {

	rawPath := strings.TrimPrefix(c.Param("path"), "/")
	key, err := parsePDFFilename(rawPath)
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
		return
	case err != nil:
		slog.Error("get result for pdf", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	pdfBytes, err := pdfgen.Generate(res)
	if err != nil {
		slog.Error("generate pdf", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "pdf generation failed"})
		return
	}

	fullPath := filepath.Join(h.pdfRoot, rawPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		slog.Error("mkdir for pdf", "err", err)

	} else if err := writeAtomic(fullPath, pdfBytes); err != nil {
		slog.Error("write pdf", "err", err)
	}

	c.Header("Content-Type", "application/pdf")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func writeAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func parsePDFFilename(p string) (result.Key, error) {
	parts := strings.Split(p, "/")
	if len(parts) != 4 {
		return result.Key{}, fmt.Errorf("malformed path")
	}
	name := strings.TrimSuffix(parts[3], ".pdf")
	fields := strings.Split(name, "-")
	if len(fields) != 4 {
		return result.Key{}, fmt.Errorf("malformed filename")
	}
	year, err := strconv.ParseInt(fields[1], 10, 16)
	if err != nil {
		return result.Key{}, fmt.Errorf("invalid year")
	}
	roll, err := strconv.ParseInt(fields[2], 10, 32)
	if err != nil {
		return result.Key{}, fmt.Errorf("invalid roll")
	}
	reg, err := strconv.ParseInt(fields[3], 10, 64)
	if err != nil {
		return result.Key{}, fmt.Errorf("invalid registration")
	}
	return result.Key{
		Exam:         "SSC",
		ExamYear:     int16(year),
		Board:        fields[0],
		Roll:         int32(roll),
		Registration: reg,
	}, nil
}

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