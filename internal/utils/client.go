package utils

import (
	"net"
	"net/http"
	"strings"
)

// GetClientIP 获取客户端真实IP地址
func GetClientIP(r *http.Request) string {
	// 尝试从X-Forwarded-For头获取IP
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For可能包含多个IP，取第一个
		if ips := strings.Split(xff, ","); len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 尝试从X-Real-IP头获取IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 尝试从X-Forwarded头获取IP
	if xf := r.Header.Get("X-Forwarded"); xf != "" {
		return xf
	}

	// 尝试从Forwarded头获取IP
	if f := r.Header.Get("Forwarded"); f != "" {
		// Forwarded头格式: for=192.0.2.60;proto=http;by=203.0.113.43
		if strings.Contains(f, "for=") {
			parts := strings.Split(f, ";")
			for _, part := range parts {
				if strings.HasPrefix(part, "for=") {
					ip := strings.TrimPrefix(part, "for=")
					return strings.Trim(ip, `"`)
				}
			}
		}
	}

	// 最后使用RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}