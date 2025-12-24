package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/zeromicro/go-zero/rest/httpx"
	"golang.org/x/time/rate"

	"github.com/zero-net-panel/zero-net-panel/internal/config"
)

// AccessMiddleware enforces IP allowlist and simple rate limiting per client.
type AccessMiddleware struct {
	allowedNets []*net.IPNet
	limiterRPS  float64
	burst       int

	limiters sync.Map // map[string]*rate.Limiter keyed by client IP
}

// NewAccessMiddleware builds middleware from admin access config.
func NewAccessMiddleware(cfg config.AdminAccessConfig) *AccessMiddleware {
	var nets []*net.IPNet
	for _, cidr := range cfg.AllowCIDRs {
		_, network, err := net.ParseCIDR(strings.TrimSpace(cidr))
		if err == nil && network != nil {
			nets = append(nets, network)
		}
	}

	rps := float64(0)
	if cfg.RateLimitPerMinute > 0 {
		rps = float64(cfg.RateLimitPerMinute) / 60.0
	}

	return &AccessMiddleware{
		allowedNets: nets,
		limiterRPS:  rps,
		burst:       cfg.Burst,
	}
}

// Handler returns the http handler middleware.
func (m *AccessMiddleware) Handler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)

		if len(m.allowedNets) > 0 && !m.ipAllowed(ip) {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusForbidden, map[string]any{
				"message": "access denied",
			})
			return
		}

		if m.limiterRPS > 0 {
			lim := m.getLimiter(ip)
			if lim != nil && !lim.Allow() {
				httpx.WriteJsonCtx(r.Context(), w, http.StatusTooManyRequests, map[string]any{
					"message": "rate limit exceeded",
				})
				return
			}
		}

		next(w, r)
	}
}

func (m *AccessMiddleware) ipAllowed(ip net.IP) bool {
	if ip == nil {
		return false
	}
	for _, network := range m.allowedNets {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func (m *AccessMiddleware) getLimiter(ip net.IP) *rate.Limiter {
	key := ip.String()
	if key == "<nil>" {
		key = "unknown"
	}

	value, ok := m.limiters.Load(key)
	if ok {
		return value.(*rate.Limiter)
	}

	lim := rate.NewLimiter(rate.Limit(m.limiterRPS), m.burst)
	m.limiters.Store(key, lim)
	return lim
}

func clientIP(r *http.Request) net.IP {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if parsed := net.ParseIP(ip); parsed != nil {
				return parsed
			}
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		if parsed := net.ParseIP(host); parsed != nil {
			return parsed
		}
	}

	return net.ParseIP(r.RemoteAddr)
}
