package handlers

import (
	"fmt"
	"net"
	"net/http"

	qrcode "github.com/skip2/go-qrcode"
)

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
			// 사설 IP 대역만 허용
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
// 현재 서버의 로컬 IP로 /mobile URL을 실제 QR 코드 PNG 이미지로 반환
func MobileQRHandler(w http.ResponseWriter, r *http.Request) {
	ip := getLocalIP()
	// 요청의 Host 헤더에서 포트 추출 (기본값 8080)
	port := "8080"
	if _, p, err := net.SplitHostPort(r.Host); err == nil && p != "" {
		port = p
	}
	targetURL := fmt.Sprintf("http://%s:%s/mobile", ip, port)

	png, err := qrcode.Encode(targetURL, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "QR 생성 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(png)
}

const manifestJSON = `{
  "name": "easyPreparation Remote",
  "short_name": "EP Remote",
  "description": "예배 슬라이드 리모컨",
  "start_url": "/mobile",
  "display": "standalone",
  "orientation": "portrait",
  "theme_color": "#002045",
  "background_color": "#f9f9ff",
  "icons": [
    {
      "src": "/mobile/icon-192.svg",
      "sizes": "192x192",
      "type": "image/svg+xml"
    }
  ]
}`

const serviceWorkerJS = `
const CACHE_NAME = 'ep-remote-v1';
const OFFLINE_URL = '/mobile';

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.add(OFFLINE_URL);
    })
  );
  self.skipWaiting();
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((keys) => {
      return Promise.all(
        keys.filter((k) => k !== CACHE_NAME).map((k) => caches.delete(k))
      );
    })
  );
  self.clients.claim();
});

self.addEventListener('fetch', (event) => {
  // POST/navigate/jump 등 API 요청은 캐시하지 않음
  if (event.request.method !== 'GET') return;
  if (event.request.url.includes('/ws')) return;
  if (event.request.url.includes('/display/')) return;

  event.respondWith(
    fetch(event.request).catch(() => {
      return caches.match(OFFLINE_URL);
    })
  );
});
`

const mobileRemoteHTML = `<!DOCTYPE html>
<html lang="ko">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
<meta name="mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-status-bar-style" content="default">
<meta name="apple-mobile-web-app-title" content="EP Remote">
<meta name="theme-color" content="#002045">
<link rel="manifest" href="/mobile/manifest.json">
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800;900&display=swap" rel="stylesheet">
<link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght,FILL@100..700,0..1&display=swap" rel="stylesheet">
<title>EP Remote</title>
<style>
  * { margin:0; padding:0; box-sizing:border-box; -webkit-tap-highlight-color:transparent; }

  :root {
    --primary: #002045;
    --secondary: #0051d5;
    --secondary-container: #316bf3;
    --surface: #f9f9ff;
    --surface-container-low: #f0f3ff;
    --surface-container: #e7eefe;
    --surface-container-high: #e2e8f8;
    --surface-container-highest: #dce2f3;
    --on-surface: #151c27;
    --on-surface-variant: #43474e;
    --outline: #74777f;
    --outline-variant: #c4c6cf;
    --error: #ba1a1a;
    --inverse-primary: #adc7f7;
  }

  html, body {
    width:100%; height:100%;
    background: var(--surface);
    color: var(--on-surface);
    font-family: 'Inter', -apple-system, sans-serif;
    overflow:hidden;
    user-select:none;
    -webkit-user-select:none;
    -webkit-font-smoothing: antialiased;
  }

  .material-symbols-outlined {
    font-variation-settings: 'FILL' 0, 'wght' 400, 'GRAD' 0, 'opsz' 24;
    font-size: 24px;
    line-height: 1;
  }
  .fill-icon { font-variation-settings: 'FILL' 1, 'wght' 400, 'GRAD' 0, 'opsz' 24; }

  /* ── 전체 레이아웃 ── */
  #app {
    display: flex;
    flex-direction: column;
    height: 100dvh;
    height: 100vh;
    max-width: 480px;
    margin: 0 auto;
    position: relative;
    overflow: hidden;
  }

  /* ── Top App Bar ── */
  #header {
    flex-shrink: 0;
    background: rgba(249,249,255,0.85);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    padding: 0 20px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 64px;
    border-bottom: 1px solid var(--outline-variant);
    z-index: 50;
  }
  #header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  #app-logo {
    font-size: 18px;
    font-weight: 900;
    color: var(--primary);
    letter-spacing: -0.03em;
    line-height: 1;
  }
  #header-right {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  #notif-btn {
    position: relative;
    cursor: pointer;
    color: var(--primary);
    display: flex;
    align-items: center;
  }
  #ws-badge {
    position: absolute;
    top: -2px;
    right: -2px;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--outline);
    border: 1.5px solid var(--surface);
    transition: background 0.3s;
  }
  #ws-badge.connected { background: #1e8c4a; }
  #ws-badge.error { background: var(--error); }

  /* ── 스크롤 영역 ── */
  #scroll-area {
    flex: 1;
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
    padding-bottom: 80px;
  }

  /* ── Live Monitor ── */
  #live-monitor {
    margin: 16px 16px 0;
    background: var(--primary);
    border-radius: 12px;
    overflow: hidden;
    box-shadow: 0 4px 20px rgba(0,32,69,0.25);
  }
  #live-monitor-inner {
    position: relative;
    background: #000;
    min-height: 120px;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    padding: 12px 16px 16px;
  }
  #live-badge-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  #live-badge {
    display: flex;
    align-items: center;
    gap: 5px;
    background: var(--error);
    color: #fff;
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    padding: 3px 8px;
    border-radius: 4px;
  }
  #live-dot {
    width: 6px; height: 6px;
    border-radius: 50%;
    background: #fff;
    animation: pulse 1.5s infinite;
  }
  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.3; }
  }
  #live-badge.offline {
    background: var(--outline);
  }
  #progress-pill {
    background: rgba(255,255,255,0.12);
    color: rgba(255,255,255,0.7);
    font-size: 10px;
    font-weight: 600;
    padding: 3px 8px;
    border-radius: 20px;
  }
  #monitor-bottom {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  #next-label {
    font-size: 9px;
    font-weight: 700;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: rgba(255,255,255,0.5);
  }
  #next-title {
    font-size: 18px;
    font-weight: 700;
    color: #fff;
    line-height: 1.2;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* ── Nav Buttons ── */
  #nav-area {
    display: flex;
    gap: 10px;
    padding: 12px 16px;
  }
  .nav-btn {
    flex: 1;
    height: 52px;
    border: none;
    border-radius: 12px;
    font-size: 13px;
    font-weight: 700;
    letter-spacing: 0.02em;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    transition: background 0.15s, transform 0.1s, box-shadow 0.15s;
    font-family: 'Inter', sans-serif;
  }
  .nav-btn:active { transform: scale(0.96); }
  #btn-prev {
    background: var(--surface-container);
    color: var(--primary);
    border: 1px solid var(--outline-variant);
  }
  #btn-prev:active { background: var(--surface-container-high); }
  #btn-next {
    background: var(--secondary);
    color: #fff;
    box-shadow: 0 2px 8px rgba(0,81,213,0.3);
    flex: 2;
  }
  #btn-next:active { background: #0042b0; }

  /* ── Quick Controls ── */
  #quick-controls {
    padding: 0 16px 12px;
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 10px;
  }
  .quick-btn {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 14px 8px;
    border-radius: 12px;
    border: none;
    cursor: pointer;
    gap: 6px;
    transition: transform 0.1s, background 0.15s;
    font-family: 'Inter', sans-serif;
  }
  .quick-btn:active { transform: scale(0.95); }
  .quick-btn-label {
    font-size: 9px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    line-height: 1.2;
    text-align: center;
  }
  #qbtn-stage {
    background: var(--secondary-container);
    color: #fff;
  }
  #qbtn-stage .material-symbols-outlined { color: #fff; }
  #qbtn-timer {
    background: #fff;
    color: var(--primary);
    border: 1px solid var(--outline-variant);
  }
  #qbtn-timer.active {
    background: #e8f5e9;
    border-color: #1e8c4a;
    color: #1e8c4a;
  }
  #qbtn-timer .material-symbols-outlined { color: inherit; }
  #qbtn-stream {
    background: #fff;
    color: var(--primary);
    border: 1px solid var(--outline-variant);
  }
  #qbtn-stream.live {
    background: #fdecea;
    border-color: var(--error);
    color: var(--error);
  }
  #qbtn-stream.starting {
    background: #fff8e1;
    border-color: #f59e0b;
    color: #b45309;
  }
  #qbtn-stream .material-symbols-outlined { color: inherit; }

  /* ── Sequence Section ── */
  #sequence-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    padding: 4px 16px 12px;
  }
  #sequence-title {
    font-size: 22px;
    font-weight: 900;
    color: var(--primary);
    letter-spacing: -0.03em;
  }
  #sequence-meta {
    font-size: 12px;
    font-weight: 600;
    color: var(--secondary);
  }

  /* ── Order List ── */
  #order-list {
    padding: 0 16px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .order-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px;
    background: var(--surface-container-low);
    border-radius: 12px;
    cursor: pointer;
    transition: background 0.1s, transform 0.1s;
    border: 1px solid transparent;
  }
  .order-item:active { transform: scale(0.99); background: var(--surface-container); }
  .order-item.active {
    background: #fff;
    border-left: 4px solid var(--secondary);
    border-radius: 12px;
    box-shadow: 0 2px 12px rgba(0,32,69,0.1);
    border-color: transparent;
    border-left-color: var(--secondary);
  }
  .item-left {
    display: flex;
    align-items: center;
    gap: 12px;
    flex: 1;
    min-width: 0;
  }
  .item-num-circle {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    background: var(--surface-container-high);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 12px;
    font-weight: 800;
    color: var(--primary);
    flex-shrink: 0;
    transition: background 0.2s, color 0.2s;
  }
  .order-item.active .item-num-circle {
    background: rgba(0,81,213,0.1);
    color: var(--secondary);
  }
  .item-text { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
  .item-badge {
    font-size: 9px;
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--on-surface-variant);
  }
  .order-item.active .item-badge { color: var(--secondary); }
  .item-title {
    font-size: 15px;
    font-weight: 700;
    color: var(--primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .item-obj {
    font-size: 11px;
    color: var(--on-surface-variant);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .item-icon {
    color: var(--outline);
    flex-shrink: 0;
    margin-left: 8px;
  }
  .order-item.active .item-icon {
    color: var(--secondary);
    font-variation-settings: 'FILL' 1, 'wght' 400, 'GRAD' 0, 'opsz' 24;
  }

  #empty-msg {
    text-align: center;
    padding: 48px 24px;
    color: var(--on-surface-variant);
    font-size: 14px;
    line-height: 2;
  }
  #empty-msg .empty-icon {
    font-size: 48px;
    color: var(--outline-variant);
    margin-bottom: 8px;
  }

  /* ── Bottom Nav Bar ── */
  #bottom-nav {
    position: fixed;
    bottom: 0;
    left: 50%;
    transform: translateX(-50%);
    width: 100%;
    max-width: 480px;
    background: rgba(255,255,255,0.9);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    border-top: 1px solid var(--outline-variant);
    border-radius: 20px 20px 0 0;
    box-shadow: 0 -4px 20px rgba(0,0,0,0.08);
    z-index: 50;
    padding-bottom: max(12px, env(safe-area-inset-bottom));
  }
  #bottom-nav-inner {
    display: flex;
    justify-content: space-around;
    align-items: center;
    padding: 10px 8px 0;
  }
  .tab-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 6px 16px;
    border-radius: 12px;
    cursor: pointer;
    gap: 2px;
    transition: background 0.15s, transform 0.1s;
    text-decoration: none;
    background: none;
    border: none;
    font-family: 'Inter', sans-serif;
  }
  .tab-item:active { transform: scale(0.9); }
  .tab-item.active {
    background: rgba(0,81,213,0.1);
    color: var(--secondary);
  }
  .tab-item.active .material-symbols-outlined {
    font-variation-settings: 'FILL' 1, 'wght' 400, 'GRAD' 0, 'opsz' 24;
    color: var(--secondary);
  }
  .tab-item .material-symbols-outlined { color: #94a3b8; }
  .tab-label {
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0.02em;
    color: inherit;
  }
  .tab-item.active .tab-label { color: var(--secondary); }
  .tab-item:not(.active) .tab-label { color: #94a3b8; }

  /* ── QR 모달 ── */
  #qr-modal {
    display: none;
    position: fixed;
    top: 0; left: 0; right: 0; bottom: 0;
    background: rgba(0,0,0,0.6);
    z-index: 100;
    align-items: center;
    justify-content: center;
  }
  #qr-modal.show { display: flex; }
  .modal-card {
    background: #fff;
    border-radius: 20px;
    padding: 28px 24px;
    max-width: 320px;
    width: 90%;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 14px;
    box-shadow: 0 8px 40px rgba(0,0,0,0.2);
  }
  .modal-title {
    font-size: 18px;
    font-weight: 800;
    color: var(--primary);
    letter-spacing: -0.02em;
  }
  .modal-subtitle {
    font-size: 13px;
    color: var(--on-surface-variant);
    text-align: center;
    line-height: 1.5;
    margin-top: -4px;
  }
  .modal-url {
    font-size: 12px;
    color: var(--secondary);
    font-family: monospace;
    word-break: break-all;
    text-align: center;
    background: var(--surface-container-low);
    padding: 10px 14px;
    border-radius: 8px;
    width: 100%;
    border: 1px solid var(--outline-variant);
  }
  .modal-img {
    border-radius: 10px;
    width: 200px;
    height: 200px;
    background: #fff;
    padding: 6px;
    image-rendering: pixelated;
    border: 1px solid var(--outline-variant);
  }
  .modal-close {
    width: 100%;
    padding: 12px;
    border: none;
    border-radius: 10px;
    background: var(--surface-container);
    color: var(--primary);
    font-size: 14px;
    font-weight: 700;
    cursor: pointer;
    font-family: 'Inter', sans-serif;
    transition: background 0.15s;
  }
  .modal-close:active { background: var(--surface-container-high); }

  /* ── 확인 모달 ── */
  #confirm-modal {
    display: none;
    position: fixed;
    top: 0; left: 0; right: 0; bottom: 0;
    background: rgba(0,0,0,0.55);
    z-index: 110;
    align-items: center;
    justify-content: center;
  }
  #confirm-modal.show { display: flex; }
  .confirm-card {
    background: #fff;
    border-radius: 20px;
    padding: 24px 20px 18px;
    max-width: 300px;
    width: 90%;
    display: flex;
    flex-direction: column;
    gap: 10px;
    box-shadow: 0 8px 40px rgba(0,0,0,0.15);
  }
  .confirm-title {
    font-size: 17px;
    font-weight: 800;
    color: var(--primary);
    letter-spacing: -0.02em;
  }
  .confirm-body {
    font-size: 13px;
    color: var(--on-surface-variant);
    line-height: 1.6;
  }
  .confirm-btns {
    display: flex;
    gap: 8px;
    margin-top: 6px;
  }
  .confirm-cancel {
    flex: 1;
    padding: 12px;
    border: 1px solid var(--outline-variant);
    border-radius: 10px;
    background: #fff;
    color: var(--on-surface-variant);
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    font-family: 'Inter', sans-serif;
  }
  .confirm-ok {
    flex: 1;
    padding: 12px;
    border: none;
    border-radius: 10px;
    background: var(--error);
    color: #fff;
    font-size: 14px;
    font-weight: 700;
    cursor: pointer;
    font-family: 'Inter', sans-serif;
  }
  .confirm-ok.start { background: var(--secondary); }

  /* ── 스와이프 힌트 ── */
  #swipe-hint {
    display: none;
    position: fixed;
    top: 0; left: 0; right: 0; bottom: 0;
    background: rgba(0,0,0,0.55);
    z-index: 99;
    align-items: center;
    justify-content: center;
  }
  #swipe-hint.show { display: flex; }
  .hint-box {
    background: #fff;
    border-radius: 20px;
    padding: 28px 24px;
    max-width: 280px;
    text-align: center;
    display: flex;
    flex-direction: column;
    gap: 12px;
    box-shadow: 0 8px 40px rgba(0,0,0,0.15);
  }
  .hint-icon { color: var(--secondary); }
  .hint-title { font-size: 17px; font-weight: 800; color: var(--primary); }
  .hint-desc { font-size: 13px; color: var(--on-surface-variant); line-height: 1.7; }
  .hint-ok {
    padding: 13px;
    border: none;
    border-radius: 10px;
    background: var(--secondary);
    color: #fff;
    font-size: 14px;
    font-weight: 700;
    cursor: pointer;
    font-family: 'Inter', sans-serif;
  }

  /* ── 토스트 ── */
  #toast {
    position: fixed;
    bottom: 90px;
    left: 50%;
    transform: translateX(-50%) translateY(16px);
    background: rgba(21,28,39,0.92);
    color: #fff;
    padding: 10px 20px;
    border-radius: 20px;
    font-size: 13px;
    font-weight: 500;
    opacity: 0;
    transition: opacity 0.25s, transform 0.25s;
    pointer-events: none;
    white-space: nowrap;
    z-index: 200;
    backdrop-filter: blur(8px);
  }
  #toast.show { opacity: 1; transform: translateX(-50%) translateY(0); }
</style>
</head>
<body>
<div id="app">
  <!-- Top App Bar -->
  <header id="header">
    <div id="header-left">
      <span class="material-symbols-outlined" style="color:var(--primary);cursor:pointer;" onclick="scrollToActive()">menu</span>
      <span id="app-logo">easyPreparation</span>
    </div>
    <div id="header-right">
      <div id="notif-btn">
        <span class="material-symbols-outlined" style="color:var(--primary);">notifications</span>
        <span id="ws-badge"></span>
      </div>
    </div>
  </header>

  <!-- 스크롤 영역 -->
  <div id="scroll-area">
    <!-- Live Monitor -->
    <div id="live-monitor">
      <div id="live-monitor-inner">
        <div id="live-badge-row">
          <div id="live-badge" class="offline">
            <span id="live-dot"></span>
            <span id="live-badge-text">OFFLINE</span>
          </div>
          <span id="progress-pill">-/-</span>
        </div>
        <div id="monitor-bottom">
          <div id="next-label">현재 항목</div>
          <div id="next-title">연결 중...</div>
        </div>
      </div>
    </div>

    <!-- Nav Buttons -->
    <div id="nav-area">
      <button class="nav-btn" id="btn-prev" onclick="navigate('prev')">
        <span class="material-symbols-outlined" style="font-size:18px;">arrow_back</span>
        <span>이전</span>
      </button>
      <button class="nav-btn" id="btn-next" onclick="navigate('next')">
        <span>다음</span>
        <span class="material-symbols-outlined" style="font-size:18px;">arrow_forward</span>
      </button>
    </div>

    <!-- Quick Controls -->
    <div id="quick-controls">
      <button class="quick-btn" id="qbtn-stage" onclick="openDisplay()">
        <span class="material-symbols-outlined fill-icon">monitor</span>
        <span class="quick-btn-label">Stage<br>Display</span>
      </button>
      <button class="quick-btn" id="qbtn-timer" onclick="toggleTimer()">
        <span class="material-symbols-outlined" id="timer-icon-ms">play_circle</span>
        <span class="quick-btn-label" id="timer-label">Auto<br>Timer</span>
      </button>
      <button class="quick-btn" id="qbtn-stream" onclick="toggleStream()">
        <span class="material-symbols-outlined" id="stream-icon-ms">radio_button_checked</span>
        <span class="quick-btn-label" id="stream-label">LIVE</span>
      </button>
    </div>

    <!-- Sequence Section -->
    <div id="sequence-header">
      <span id="sequence-title">Sequence</span>
      <span id="sequence-meta"></span>
    </div>

    <!-- Order List -->
    <div id="order-list"></div>
    <div id="empty-msg" style="display:none;">
      <div class="material-symbols-outlined empty-icon">queue_music</div>
      <div>예배 순서가 없습니다.<br>주보 탭에서 순서를 전송해주세요.</div>
    </div>

    <!-- 하단 여백 -->
    <div style="height:16px;"></div>
  </div>

  <!-- Bottom Nav Bar -->
  <nav id="bottom-nav">
    <div id="bottom-nav-inner">
      <button class="tab-item active" id="tab-sequence" onclick="switchTab('sequence')">
        <span class="material-symbols-outlined">queue_music</span>
        <span class="tab-label">Sequence</span>
      </button>
      <button class="tab-item" id="tab-live" onclick="switchTab('live')">
        <span class="material-symbols-outlined">radio</span>
        <span class="tab-label">Live</span>
      </button>
      <button class="tab-item" id="tab-controls" onclick="showQR()">
        <span class="material-symbols-outlined">settings_remote</span>
        <span class="tab-label">Controls</span>
      </button>
      <button class="tab-item" id="tab-settings" onclick="switchTab('settings')">
        <span class="material-symbols-outlined">settings</span>
        <span class="tab-label">Settings</span>
      </button>
    </div>
  </nav>
</div>

<!-- QR 모달 -->
<div id="qr-modal">
  <div class="modal-card">
    <div class="modal-title">모바일 접속 주소</div>
    <div class="modal-subtitle">같은 WiFi에 연결된 기기에서<br>스캔하면 리모컨을 사용할 수 있습니다</div>
    <img class="modal-img" id="qr-img" src="/mobile/qr.png" alt="QR">
    <div class="modal-url" id="qr-url"></div>
    <button class="modal-close" onclick="hideQR()">닫기</button>
  </div>
</div>

<!-- 스트리밍 확인 모달 -->
<div id="confirm-modal">
  <div class="confirm-card">
    <div class="confirm-title" id="confirm-title">확인</div>
    <div class="confirm-body" id="confirm-body"></div>
    <div class="confirm-btns">
      <button class="confirm-cancel" onclick="hideConfirm()">취소</button>
      <button class="confirm-ok" id="confirm-ok-btn" onclick="confirmAction()">확인</button>
    </div>
  </div>
</div>

<!-- 스와이프 힌트 -->
<div id="swipe-hint">
  <div class="hint-box">
    <span class="material-symbols-outlined hint-icon" style="font-size:40px;">swipe_left</span>
    <div class="hint-title">슬라이드 제어</div>
    <div class="hint-desc">
      좌우로 스와이프하거나<br>
      이전 / 다음 버튼으로<br>
      슬라이드를 이동하세요.
    </div>
    <button class="hint-ok" onclick="dismissHint()">시작하기</button>
  </div>
</div>

<!-- 토스트 -->
<div id="toast"></div>

<script>
'use strict';

// ── 상태 ──
let items = [];
let currentIdx = 0;
let timerEnabled = false;
let streamLive = false;
let streamBusy = false;
let ws = null;
let wsReconnectTimer = null;
let touchStartX = 0;
let touchStartY = 0;
let touchMoved = false;
let pendingConfirmAction = null;

// ── WebSocket 연결 ──
function connectWS() {
  const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
  ws = new WebSocket(protocol + '//' + location.host + '/ws');

  ws.onopen = function() {
    setWsDot('connected');
    if (wsReconnectTimer) { clearTimeout(wsReconnectTimer); wsReconnectTimer = null; }
  };

  ws.onmessage = function(e) {
    try {
      const msg = JSON.parse(e.data);
      handleWS(msg);
    } catch(_) {}
  };

  ws.onclose = function() {
    setWsDot('error');
    wsReconnectTimer = setTimeout(connectWS, 3000);
  };

  ws.onerror = function() {
    setWsDot('error');
    ws.close();
  };
}

function handleWS(msg) {
  switch (msg.type) {
    case 'order':
      items = msg.items || [];
      if (typeof msg.idx === 'number') currentIdx = msg.idx;
      renderList();
      updateMonitor();
      break;

    case 'position':
      if (typeof msg.idx === 'number') {
        currentIdx = msg.idx;
        updateMonitor();
        highlightCurrent();
        scrollToActive();
      }
      break;

    case 'navigate':
      if (msg.direction === 'jump' && typeof msg.idx === 'number') {
        currentIdx = msg.idx;
        updateMonitor();
        highlightCurrent();
        scrollToActive();
      }
      break;

    case 'timer_state':
      timerEnabled = !!msg.enabled;
      updateTimerBtn();
      break;

    case 'keepalive':
      break;
  }
}

function setWsDot(state) {
  const badge = document.getElementById('ws-badge');
  badge.className = state;
}

// ── 네비게이션 ──
function navigate(dir) {
  fetch('/display/navigate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ direction: dir })
  }).catch(function() { showToast('연결 오류'); });
}

function jumpTo(idx) {
  fetch('/display/jump', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ index: idx })
  }).catch(function() { showToast('연결 오류'); });
}

function openDisplay() {
  window.open('/display', '_blank');
}

// ── 탭 전환 (시각적 토글만) ──
function switchTab(tab) {
  const tabs = ['sequence', 'live', 'controls', 'settings'];
  tabs.forEach(function(t) {
    const el = document.getElementById('tab-' + t);
    if (el) el.classList.remove('active');
  });
  const active = document.getElementById('tab-' + tab);
  if (active) active.classList.add('active');
  if (tab === 'live') { toggleStream(); }
}

// ── 타이머 제어 ──
function toggleTimer() {
  fetch('/display/timer', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ action: 'toggle' })
  }).then(function() {
    timerEnabled = !timerEnabled;
    updateTimerBtn();
  }).catch(function() { showToast('연결 오류'); });
}

function updateTimerBtn() {
  const btn = document.getElementById('qbtn-timer');
  const icon = document.getElementById('timer-icon-ms');
  const label = document.getElementById('timer-label');
  if (timerEnabled) {
    btn.classList.add('active');
    icon.textContent = 'pause_circle';
    label.textContent = 'Timer\nON';
  } else {
    btn.classList.remove('active');
    icon.textContent = 'play_circle';
    label.textContent = 'Auto\nTimer';
  }
}

// ── 스트리밍 제어 ──
function toggleStream() {
  if (streamBusy) { showToast('처리 중...'); return; }
  if (streamLive) {
    showConfirm(
      '방송 종료',
      'OBS 스트리밍을 종료하시겠습니까?',
      false,
      doStopStream
    );
  } else {
    showConfirm(
      '방송 시작',
      'OBS 스트리밍을 시작하시겠습니까?',
      true,
      doStartStream
    );
  }
}

function doStartStream() {
  streamBusy = true;
  updateStreamBtn('starting');
  fetch('/api/schedule/stream', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ action: 'start' })
  })
  .then(function(r) { return r.json(); })
  .then(function(data) {
    streamBusy = false;
    if (data.ok || data.active) {
      streamLive = true;
      showToast('방송이 시작되었습니다');
    } else {
      showToast('방송 시작 실패: ' + (data.error || '알 수 없는 오류'));
    }
    updateStreamBtn('');
    updateLiveBadge();
  })
  .catch(function() {
    streamBusy = false;
    updateStreamBtn('');
    showToast('연결 오류');
  });
}

function doStopStream() {
  streamBusy = true;
  updateStreamBtn('starting');
  fetch('/api/schedule/stream', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ action: 'stop' })
  })
  .then(function(r) { return r.json(); })
  .then(function(data) {
    streamBusy = false;
    streamLive = false;
    showToast('방송이 종료되었습니다');
    updateStreamBtn('');
    updateLiveBadge();
  })
  .catch(function() {
    streamBusy = false;
    updateStreamBtn('');
    showToast('연결 오류');
  });
}

function updateStreamBtn(forceCls) {
  const btn = document.getElementById('qbtn-stream');
  const icon = document.getElementById('stream-icon-ms');
  const label = document.getElementById('stream-label');
  btn.className = 'quick-btn';
  if (forceCls === 'starting') {
    btn.classList.add('starting');
    icon.textContent = 'hourglass_empty';
    label.textContent = '처리 중';
  } else if (streamLive) {
    btn.classList.add('live');
    icon.textContent = 'radio_button_checked';
    label.textContent = 'LIVE';
  } else {
    icon.textContent = 'radio_button_checked';
    label.textContent = 'LIVE';
  }
}

function updateLiveBadge() {
  const badge = document.getElementById('live-badge');
  const text = document.getElementById('live-badge-text');
  if (streamLive) {
    badge.classList.remove('offline');
    text.textContent = 'LIVE';
  } else {
    badge.classList.add('offline');
    text.textContent = 'OFFLINE';
  }
}

// ── 확인 모달 ──
function showConfirm(title, body, isStart, onOk) {
  document.getElementById('confirm-title').textContent = title;
  document.getElementById('confirm-body').textContent = body;
  const okBtn = document.getElementById('confirm-ok-btn');
  okBtn.textContent = isStart ? '시작' : '종료';
  okBtn.className = 'confirm-ok' + (isStart ? ' start' : '');
  pendingConfirmAction = onOk;
  document.getElementById('confirm-modal').classList.add('show');
}

function hideConfirm() {
  pendingConfirmAction = null;
  document.getElementById('confirm-modal').classList.remove('show');
}

function confirmAction() {
  var action = pendingConfirmAction;
  hideConfirm();
  if (action) action();
}

// ── 렌더링 ──
function getBadgeText(i) {
  if (i === currentIdx) return 'Current';
  if (i === currentIdx + 1) return 'Next';
  return '';
}

function renderList() {
  const list = document.getElementById('order-list');
  const empty = document.getElementById('empty-msg');

  if (!items || items.length === 0) {
    list.innerHTML = '';
    empty.style.display = 'block';
    updateMonitor();
    return;
  }
  empty.style.display = 'none';

  list.innerHTML = items.map(function(item, i) {
    const title = item.title || '';
    const obj = item.obj || '';
    const isActive = i === currentIdx;
    const badge = getBadgeText(i);
    const numStr = String(i + 1).padStart(2, '0');
    return '<div class="order-item' + (isActive ? ' active' : '') + '" onclick="jumpTo(' + i + ')" data-idx="' + i + '">' +
      '<div class="item-left">' +
        '<div class="item-num-circle">' + numStr + '</div>' +
        '<div class="item-text">' +
          (badge ? '<span class="item-badge">' + badge + '</span>' : '<span class="item-badge">&nbsp;</span>') +
          '<span class="item-title">' + escHtml(title) + '</span>' +
          (obj ? '<span class="item-obj">' + escHtml(obj) + '</span>' : '') +
        '</div>' +
      '</div>' +
      '<span class="material-symbols-outlined item-icon">' + (isActive ? 'equalizer' : 'drag_indicator') + '</span>' +
    '</div>';
  }).join('');
}

function highlightCurrent() {
  const allItems = document.querySelectorAll('.order-item');
  allItems.forEach(function(el) {
    const idx = parseInt(el.getAttribute('data-idx'), 10);
    const isActive = idx === currentIdx;
    if (isActive) {
      el.classList.add('active');
      el.querySelector('.item-icon').textContent = 'equalizer';
    } else {
      el.classList.remove('active');
      el.querySelector('.item-icon').textContent = 'drag_indicator';
    }
    const badge = el.querySelector('.item-badge');
    if (badge) {
      const b = getBadgeText(idx);
      badge.textContent = b || '\u00a0';
    }
  });
}

function scrollToActive() {
  const active = document.querySelector('.order-item.active');
  if (active) {
    active.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }
}

function updateMonitor() {
  const nextTitle = document.getElementById('next-title');
  const progressPill = document.getElementById('progress-pill');
  const nextLabel = document.getElementById('next-label');
  const seqMeta = document.getElementById('sequence-meta');

  if (items && items.length > 0 && currentIdx >= 0 && currentIdx < items.length) {
    nextTitle.textContent = items[currentIdx].title || '';
    progressPill.textContent = (currentIdx + 1) + ' / ' + items.length;
    nextLabel.textContent = '현재 항목';
    seqMeta.textContent = (currentIdx + 1) + ' / ' + items.length;
  } else {
    nextTitle.textContent = '순서 없음';
    progressPill.textContent = '-/-';
    nextLabel.textContent = '현재 항목';
    seqMeta.textContent = '';
  }
}

// ── 스트리밍 상태 폴링 ──
function pollStreamStatus() {
  fetch('/display/status')
    .then(function(r) { return r.json(); })
    .then(function(data) {
      const live = !!(data.stream && data.stream.active);
      if (!streamBusy && live !== streamLive) {
        streamLive = live;
        updateStreamBtn('');
        updateLiveBadge();
      }
    })
    .catch(function() {});
}

// ── QR 모달 ──
function showQR() {
  const modal = document.getElementById('qr-modal');
  const urlEl = document.getElementById('qr-url');
  const mobileURL = location.protocol + '//' + location.host + '/mobile';
  urlEl.textContent = mobileURL;
  document.getElementById('qr-img').src = '/mobile/qr.png?t=' + Date.now();
  modal.classList.add('show');
}

function hideQR() {
  document.getElementById('qr-modal').classList.remove('show');
}

// ── 토스트 ──
let toastTimer = null;
function showToast(msg) {
  const el = document.getElementById('toast');
  el.textContent = msg;
  el.classList.add('show');
  if (toastTimer) clearTimeout(toastTimer);
  toastTimer = setTimeout(function() { el.classList.remove('show'); }, 2200);
}

// ── 터치 스와이프 ──
document.addEventListener('touchstart', function(e) {
  touchStartX = e.touches[0].clientX;
  touchStartY = e.touches[0].clientY;
  touchMoved = false;
}, { passive: true });

document.addEventListener('touchmove', function(e) {
  const dx = Math.abs(e.touches[0].clientX - touchStartX);
  const dy = Math.abs(e.touches[0].clientY - touchStartY);
  if (dx > 10 || dy > 10) touchMoved = true;
}, { passive: true });

document.addEventListener('touchend', function(e) {
  if (!touchMoved) return;
  const dx = e.changedTouches[0].clientX - touchStartX;
  const dy = e.changedTouches[0].clientY - touchStartY;

  if (Math.abs(dy) > Math.abs(dx)) return;
  if (Math.abs(dx) < 60) return;

  const wrap = document.getElementById('scroll-area');
  const rect = wrap.getBoundingClientRect();
  const startY = e.changedTouches[0].clientY;
  if (startY >= rect.top && startY <= rect.bottom) return;

  if (dx < 0) {
    navigate('next');
    showToast('다음 슬라이드');
  } else {
    navigate('prev');
    showToast('이전 슬라이드');
  }
}, { passive: true });

// ── 키보드 지원 ──
document.addEventListener('keydown', function(e) {
  if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return;
  if (e.key === 'ArrowRight' || e.key === ' ') { e.preventDefault(); navigate('next'); }
  if (e.key === 'ArrowLeft') { e.preventDefault(); navigate('prev'); }
  if (e.key === 'Escape') { hideQR(); hideConfirm(); }
});

// ── 첫 방문 힌트 ──
function dismissHint() {
  document.getElementById('swipe-hint').classList.remove('show');
  try { localStorage.setItem('ep_remote_hint_seen', '1'); } catch(_) {}
}

// ── PWA Service Worker 등록 ──
if ('serviceWorker' in navigator) {
  navigator.serviceWorker.register('/mobile/sw.js').catch(function() {});
}

// ── HTML 이스케이프 ──
function escHtml(s) {
  return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

// ── 초기화 ──
(function init() {
  connectWS();
  pollStreamStatus();
  setInterval(pollStreamStatus, 10000);

  try {
    if (!localStorage.getItem('ep_remote_hint_seen')) {
      document.getElementById('swipe-hint').classList.add('show');
    }
  } catch(_) {}
})();
</script>
</body>
</html>`
