package middleware

import (
	"net"
	"net/http"
	"strings"
)

// isAllowedOrigin — localhost, 127.0.0.1, 사설 IP(192.168.x.x, 10.x.x.x, 172.16-31.x.x)만 허용
func isAllowedOrigin(origin string) bool {
	if origin == "" {
		return false
	}

	// http:// 또는 https:// 프로토콜 제거 후 호스트 추출
	host := origin
	if idx := strings.Index(host, "://"); idx != -1 {
		host = host[idx+3:]
	}
	// 포트 제거 (호스트:포트 형태)
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	// localhost, 127.0.0.1, Wails WebView 허용
	if host == "localhost" || host == "127.0.0.1" || host == "::1" || host == "wails.localhost" {
		return true
	}

	// 사설 IP 대역 확인
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	privateRanges := []struct {
		network string
		mask    string
	}{
		{"10.0.0.0", "255.0.0.0"},         // 10.0.0.0/8
		{"172.16.0.0", "255.240.0.0"},      // 172.16.0.0/12
		{"192.168.0.0", "255.255.0.0"},     // 192.168.0.0/16
	}
	for _, r := range privateRanges {
		network := net.ParseIP(r.network)
		mask := net.IPMask(net.ParseIP(r.mask).To4())
		if network != nil && mask != nil {
			subnet := &net.IPNet{IP: network, Mask: mask}
			if subnet.Contains(ip) {
				return true
			}
		}
	}

	return false
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if isAllowedOrigin(origin) {
			// 허용된 origin을 반사 (reflect)
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		// 허용되지 않은 origin이면 Access-Control-Allow-Origin 헤더를 설정하지 않음
		// → 브라우저가 CORS 차단

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Preflight 요청 처리
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
