package handlers

import (
	_ "embed"
	"fmt"
	"net"
	"net/http"

	qrcode "github.com/skip2/go-qrcode"
)

//go:embed html/mobile-manifest.json
var manifestJSON string

//go:embed html/mobile-sw.js
var serviceWorkerJS string

//go:embed html/mobile-remote.html
var mobileRemoteHTML string

// getLocalIP — 로컬 WiFi/LAN IP 감지 (192.168.x.x, 10.x.x.x, 172.16-31.x.x)
func getLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "localhost"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue
			}
			// 사설 IP 대역만 허용 (192.x.x.x / 10.x.x.x / 172.16-31.x.x)
			if ip[0] == 192 || ip[0] == 10 || (ip[0] == 172 && ip[1] >= 16 && ip[1] <= 31) {
				return ip.String()
			}
		}
	}
	return "localhost"
}

// MobileRemoteHandler — GET /mobile
// 모바일 PWA 리모컨 HTML 서빙
func MobileRemoteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(mobileRemoteHTML))
}

// MobileManifestHandler — GET /mobile/manifest.json
func MobileManifestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write([]byte(manifestJSON))
}

// MobileServiceWorkerHandler — GET /mobile/sw.js
func MobileServiceWorkerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(serviceWorkerJS))
}

// MobileIconHandler — GET /mobile/icon-192.svg
func MobileIconHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=604800")
	w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="192" height="192" viewBox="0 0 192 192">
<rect width="192" height="192" rx="32" fill="#002045"/>
<text x="96" y="120" text-anchor="middle" font-size="80" font-weight="bold" fill="#adc7f7" font-family="sans-serif">EP</text>
</svg>`))
}

// MobileQRHandler — GET /mobile/qr.png
// 터널 URL이 있으면 그걸 사용, 없으면 로컬 IP로 /mobile URL QR 생성
func MobileQRHandler(w http.ResponseWriter, r *http.Request) {
	var targetURL string

	// 터널 URL 우선
	if tunnelURL := GetTunnelURL(); tunnelURL != "" {
		targetURL = tunnelURL + "/mobile"
	} else {
		ip := getLocalIP()
		// 실제 TCP 리스너 포트 추출 (Host 헤더 대신 LocalAddr 사용 — Wails WebView는 Host가 다름)
		port := "8080"
		if localAddr, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr); ok {
			if _, p, err := net.SplitHostPort(localAddr.String()); err == nil && p != "" {
				port = p
			}
		}
		targetURL = fmt.Sprintf("http://%s:%s/mobile", ip, port)
	}

	png, err := qrcode.Encode(targetURL, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "QR 생성 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(png)
}
