package middleware

import (
	"errors"
	"log/slog"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/zishan044/education-board-result-publishing-system/internal/cache"
)

type IPFilter struct {
	blocked []*net.IPNet
	allowed []*net.IPNet
	cache   *cache.Cache
}

func NewIPFilter(blockCIDRs, allowCIDRs []string, ch *cache.Cache) (*IPFilter, error) {
	f := &IPFilter{cache: ch}
	var err error
	if f.blocked, err = parseCIDRs(blockCIDRs); err != nil {
		return nil, err
	}
	if f.allowed, err = parseCIDRs(allowCIDRs); err != nil {
		return nil, err
	}
	return f, nil
}

func parseCIDRs(raw []string) ([]*net.IPNet, error) {
	nets := make([]*net.IPNet, 0, len(raw))
	for _, s := range raw {
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			ip := net.ParseIP(s)
			if ip == nil {
				return nil, err
			}
			bits := "32"
			if ip.To4() == nil {
				bits = "128"
			}
			_, n, err = net.ParseCIDR(ip.String() + "/" + bits)
			if err != nil {
				return nil, err
			}
		}
		nets = append(nets, n)
	}
	return nets, nil
}

func containsIP(nets []*net.IPNet, ip net.IP) bool {
	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

func (f *IPFilter) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ipStr := c.ClientIP()
		ip := net.ParseIP(ipStr)
		if ip == nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if len(f.allowed) > 0 && !containsIP(f.allowed, ip) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		if containsIP(f.blocked, ip) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		blocked, err := f.cache.IsIPBlocked(c.Request.Context(), ipStr)
		if err != nil && !errors.Is(err, redis.Nil) {
			// Fail open: a Valkey outage must not take down lookups.
			slog.Warn("ip block check failed, allowing request", "err", err)
		} else if blocked {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}