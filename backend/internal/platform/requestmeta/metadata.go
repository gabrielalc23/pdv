package requestmeta

import (
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
)

const maxUserAgentLength = 512

type RequestMetadata struct {
	ClientIP  string
	UserAgent string
	RequestID string
}

type Resolver struct {
	trustedCIDRs []*net.IPNet
}

func NewResolver(cidrs []string) (*Resolver, error) {
	parsed := make([]*net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		_, network, err := net.ParseCIDR(c)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, network)
	}
	return &Resolver{trustedCIDRs: parsed}, nil
}

func (r *Resolver) Extract(req *http.Request) RequestMetadata {
	clientIP := r.resolveClientIP(req)
	userAgent := truncateString(req.UserAgent(), maxUserAgentLength)
	requestID := middleware.GetReqID(req.Context())

	return RequestMetadata{
		ClientIP:  clientIP,
		UserAgent: userAgent,
		RequestID: requestID,
	}
}

func (r *Resolver) resolveClientIP(req *http.Request) string {
	remoteAddr := req.RemoteAddr
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}

	remoteIP := net.ParseIP(host)
	if remoteIP == nil {
		return host
	}

	if !r.isTrusted(remoteIP) {
		return remoteIP.String()
	}

	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		for i := len(parts) - 1; i >= 0; i-- {
			ip := net.ParseIP(strings.TrimSpace(parts[i]))
			if ip == nil {
				continue
			}
			if !r.isTrusted(ip) {
				return ip.String()
			}
		}
	}

	if xri := req.Header.Get("X-Real-IP"); xri != "" {
		if ip := net.ParseIP(strings.TrimSpace(xri)); ip != nil && !r.isTrusted(ip) {
			return ip.String()
		}
	}

	return remoteIP.String()
}

func (r *Resolver) isTrusted(ip net.IP) bool {
	for _, network := range r.trustedCIDRs {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func (r *Resolver) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		meta := r.Extract(req)
		ctx := req.Context()
		ctx = withMetadata(ctx, meta)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
