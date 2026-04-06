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
<rect width="192" height="192" rx="32" fill="#1a1a1a"/>
<text x="96" y="120" text-anchor="middle" font-size="80" font-weight="bold" fill="#4fc3f7" font-family="sans-serif">EP</text>
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
  "theme_color": "#1a1a1a",
  "background_color": "#1a1a1a",
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
<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">
<meta name="apple-mobile-web-app-title" content="EP Remote">
<meta name="theme-color" content="#1a1a1a">
<link rel="manifest" href="/mobile/manifest.json">
<title>EP Remote</title>
<style>
  * { margin:0; padding:0; box-sizing:border-box; -webkit-tap-highlight-color:transparent; }
  html, body {
    width:100%; height:100%;
    background:#1a1a1a;
    color:#ffffff;
    font-family:-apple-system, 'Apple SD Gothic Neo', 'Malgun Gothic', sans-serif;
    overflow:hidden;
    user-select:none;
    -webkit-user-select:none;
  }

  /* 전체 레이아웃 — flex column */
  #app {
    display:flex;
    flex-direction:column;
    height:100dvh;
    height:100vh;
    max-width:480px;
    margin:0 auto;
  }

  /* 헤더 */
  #header {
    flex-shrink:0;
    background:#111;
    padding:12px 16px 10px;
    border-bottom:1px solid #2a2a2a;
    display:flex;
    align-items:center;
    justify-content:space-between;
    min-height:60px;
  }
  #header-left { display:flex; flex-direction:column; gap:2px; }
  #app-name {
    font-size:13px;
    color:#888;
    letter-spacing:0.05em;
    font-weight:500;
  }
  #current-title {
    font-size:17px;
    font-weight:700;
    color:#fff;
    white-space:nowrap;
    overflow:hidden;
    text-overflow:ellipsis;
    max-width:220px;
  }
  #progress-text {
    font-size:12px;
    color:#666;
    flex-shrink:0;
  }

  /* WS 연결 상태 도트 */
  #ws-dot {
    width:8px; height:8px;
    border-radius:50%;
    background:#444;
    display:inline-block;
    margin-right:6px;
    transition:background 0.3s;
  }
  #ws-dot.connected { background:#4caf50; }
  #ws-dot.error { background:#f44336; }

  /* 네비게이션 버튼 영역 */
  #nav-area {
    flex-shrink:0;
    display:flex;
    gap:10px;
    padding:12px 16px;
    background:#181818;
    border-bottom:1px solid #2a2a2a;
  }
  .nav-btn {
    flex:1;
    height:64px;
    border:none;
    border-radius:10px;
    font-size:22px;
    font-weight:700;
    cursor:pointer;
    display:flex;
    align-items:center;
    justify-content:center;
    gap:8px;
    transition:background 0.15s, transform 0.1s;
    -webkit-tap-highlight-color:transparent;
  }
  .nav-btn:active { transform:scale(0.96); }
  #btn-prev {
    background:#2a2a2a;
    color:#ccc;
  }
  #btn-prev:active { background:#333; }
  #btn-next {
    background:#1976d2;
    color:#fff;
  }
  #btn-next:active { background:#1565c0; }
  .nav-label {
    font-size:13px;
    font-weight:600;
    letter-spacing:0.03em;
  }

  /* 예배 순서 리스트 */
  #order-list-wrap {
    flex:1;
    overflow-y:auto;
    -webkit-overflow-scrolling:touch;
    padding:8px 0;
    background:#1a1a1a;
  }
  .order-item {
    display:flex;
    align-items:center;
    padding:13px 16px;
    border-bottom:1px solid #222;
    cursor:pointer;
    transition:background 0.1s;
    min-height:52px;
    gap:12px;
  }
  .order-item:active { background:#252525; }
  .order-item.active {
    background:#0d2744;
    border-left:3px solid #1976d2;
    padding-left:13px;
  }
  .order-item.active .item-title {
    color:#4fc3f7;
    font-weight:700;
  }
  .item-num {
    font-size:12px;
    color:#555;
    width:22px;
    flex-shrink:0;
    text-align:right;
  }
  .item-dot {
    width:6px; height:6px;
    border-radius:50%;
    background:#333;
    flex-shrink:0;
    transition:background 0.2s;
  }
  .order-item.active .item-dot { background:#1976d2; }
  .item-title {
    flex:1;
    font-size:15px;
    color:#ccc;
    white-space:nowrap;
    overflow:hidden;
    text-overflow:ellipsis;
  }
  .item-obj {
    font-size:12px;
    color:#666;
    flex-shrink:0;
    max-width:90px;
    white-space:nowrap;
    overflow:hidden;
    text-overflow:ellipsis;
  }
  #empty-msg {
    text-align:center;
    padding:48px 16px;
    color:#444;
    font-size:14px;
    line-height:2;
  }

  /* 하단 컨트롤 바 */
  #bottom-bar {
    flex-shrink:0;
    display:flex;
    align-items:center;
    padding:10px 16px;
    background:#111;
    border-top:1px solid #2a2a2a;
    gap:10px;
    min-height:58px;
    padding-bottom:max(10px, env(safe-area-inset-bottom));
  }
  .ctrl-btn {
    flex:1;
    height:40px;
    border:none;
    border-radius:8px;
    font-size:13px;
    font-weight:600;
    cursor:pointer;
    display:flex;
    align-items:center;
    justify-content:center;
    gap:5px;
    transition:background 0.15s, transform 0.1s;
  }
  .ctrl-btn:active { transform:scale(0.96); }
  #btn-timer {
    background:#2a2a2a;
    color:#aaa;
  }
  #btn-timer.active {
    background:#1b3a1b;
    color:#81c784;
  }
  #btn-timer:active { background:#333; }
  #btn-stream {
    background:#2a2a2a;
    color:#888;
  }
  #btn-stream.live {
    background:#3b1b1b;
    color:#ef5350;
  }
  #btn-stream.starting {
    background:#2a2010;
    color:#ffb74d;
  }
  #btn-stream:active { background:#333; }
  #btn-qr {
    width:40px;
    height:40px;
    flex:none;
    background:#2a2a2a;
    color:#888;
    border:none;
    border-radius:8px;
    cursor:pointer;
    display:flex;
    align-items:center;
    justify-content:center;
    font-size:16px;
    transition:background 0.15s;
  }
  #btn-qr:active { background:#333; transform:scale(0.96); }

  /* QR 모달 */
  #qr-modal {
    display:none;
    position:fixed;
    top:0; left:0; right:0; bottom:0;
    background:rgba(0,0,0,0.8);
    z-index:100;
    align-items:center;
    justify-content:center;
    flex-direction:column;
    gap:16px;
  }
  #qr-modal.show { display:flex; }
  #qr-modal .modal-card {
    background:#222;
    border-radius:16px;
    padding:24px;
    max-width:320px;
    width:90%;
    display:flex;
    flex-direction:column;
    align-items:center;
    gap:12px;
  }
  #qr-modal .modal-title {
    font-size:16px;
    font-weight:700;
    color:#fff;
  }
  #qr-modal .modal-subtitle {
    font-size:12px;
    color:#888;
    text-align:center;
  }
  #qr-modal .modal-url {
    font-size:12px;
    color:#4fc3f7;
    font-family:monospace;
    word-break:break-all;
    text-align:center;
    background:#1a1a1a;
    padding:8px 12px;
    border-radius:6px;
    width:100%;
  }
  #qr-modal .modal-img {
    border-radius:8px;
    width:200px;
    height:200px;
    background:#fff;
    padding:8px;
    image-rendering:pixelated;
  }
  #qr-modal .modal-close {
    margin-top:4px;
    padding:10px 32px;
    border:none;
    border-radius:8px;
    background:#333;
    color:#ccc;
    font-size:14px;
    cursor:pointer;
  }

  /* 스와이프 힌트 오버레이 (첫 방문) */
  #swipe-hint {
    display:none;
    position:fixed;
    top:0; left:0; right:0; bottom:0;
    background:rgba(0,0,0,0.6);
    z-index:99;
    align-items:center;
    justify-content:center;
  }
  #swipe-hint.show { display:flex; }
  #swipe-hint .hint-box {
    background:#222;
    border-radius:16px;
    padding:28px 24px;
    max-width:280px;
    text-align:center;
    display:flex;
    flex-direction:column;
    gap:12px;
  }
  #swipe-hint .hint-title { font-size:16px; font-weight:700; }
  #swipe-hint .hint-desc { font-size:13px; color:#999; line-height:1.7; }
  #swipe-hint .hint-ok {
    padding:12px;
    border:none;
    border-radius:8px;
    background:#1976d2;
    color:#fff;
    font-size:14px;
    font-weight:600;
    cursor:pointer;
  }

  /* 확인 대화상자 */
  #confirm-modal {
    display:none;
    position:fixed;
    top:0; left:0; right:0; bottom:0;
    background:rgba(0,0,0,0.75);
    z-index:110;
    align-items:center;
    justify-content:center;
  }
  #confirm-modal.show { display:flex; }
  #confirm-modal .confirm-card {
    background:#242424;
    border-radius:14px;
    padding:24px 20px 16px;
    max-width:300px;
    width:90%;
    display:flex;
    flex-direction:column;
    gap:10px;
  }
  #confirm-modal .confirm-title {
    font-size:16px;
    font-weight:700;
    color:#fff;
  }
  #confirm-modal .confirm-body {
    font-size:13px;
    color:#aaa;
    line-height:1.6;
  }
  #confirm-modal .confirm-btns {
    display:flex;
    gap:8px;
    margin-top:6px;
  }
  #confirm-modal .confirm-cancel {
    flex:1;
    padding:11px;
    border:none;
    border-radius:8px;
    background:#333;
    color:#aaa;
    font-size:14px;
    cursor:pointer;
  }
  #confirm-modal .confirm-ok {
    flex:1;
    padding:11px;
    border:none;
    border-radius:8px;
    background:#c62828;
    color:#fff;
    font-size:14px;
    font-weight:600;
    cursor:pointer;
  }
  #confirm-modal .confirm-ok.start {
    background:#1565c0;
  }

  /* 토스트 알림 */
  #toast {
    position:fixed;
    bottom:80px;
    left:50%;
    transform:translateX(-50%) translateY(20px);
    background:rgba(50,50,50,0.95);
    color:#fff;
    padding:10px 20px;
    border-radius:20px;
    font-size:13px;
    opacity:0;
    transition:opacity 0.25s, transform 0.25s;
    pointer-events:none;
    white-space:nowrap;
    z-index:200;
  }
  #toast.show { opacity:1; transform:translateX(-50%) translateY(0); }
</style>
</head>
<body>
<div id="app">
  <!-- 헤더 -->
  <div id="header">
    <div id="header-left">
      <div id="app-name"><span id="ws-dot"></span>easyPreparation</div>
      <div id="current-title">연결 중...</div>
    </div>
    <div id="progress-text">-/-</div>
  </div>

  <!-- 네비게이션 버튼 -->
  <div id="nav-area">
    <button class="nav-btn" id="btn-prev" onclick="navigate('prev')">
      <span>&#9664;</span>
      <span class="nav-label">이전</span>
    </button>
    <button class="nav-btn" id="btn-next" onclick="navigate('next')">
      <span class="nav-label">다음</span>
      <span>&#9654;</span>
    </button>
  </div>

  <!-- 순서 리스트 -->
  <div id="order-list-wrap">
    <div id="order-list"></div>
    <div id="empty-msg" style="display:none;">
      예배 순서가 없습니다.<br>주보 탭에서 순서를 전송해주세요.
    </div>
  </div>

  <!-- 하단 컨트롤 -->
  <div id="bottom-bar">
    <button class="ctrl-btn" id="btn-timer" onclick="toggleTimer()">
      <span id="timer-icon">&#9654;</span>
      <span id="timer-label">타이머</span>
    </button>
    <button class="ctrl-btn" id="btn-stream" onclick="toggleStream()">
      <span id="stream-dot" style="width:8px;height:8px;border-radius:50%;background:#555;display:inline-block;flex-shrink:0;"></span>
      <span id="stream-label">LIVE</span>
    </button>
    <button id="btn-qr" onclick="showQR()" title="QR 코드">
      <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
        <path d="M3 3h7v7H3V3zm2 2v3h3V5H5zm9-2h7v7h-7V3zm2 2v3h3V5h-3zM3 14h7v7H3v-7zm2 2v3h3v-3H5zm11-2h2v2h-2v-2zm2 2h2v2h-2v-2zm-2 2h2v2h-2v-2zm2 2h2v2h-2v-2zm-4-4h2v2h-2v-2zm2 2h2v2h-2v-2z"/>
      </svg>
    </button>
  </div>
</div>

<!-- QR 모달 -->
<div id="qr-modal">
  <div class="modal-card">
    <div class="modal-title">모바일 접속 주소</div>
    <div class="modal-subtitle">같은 WiFi에 연결된 기기에서 스캔하세요</div>
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
    <div class="hint-title">슬라이드 제어</div>
    <div class="hint-desc">
      좌우로 스와이프하거나<br>
      이전 / 다음 버튼으로<br>
      슬라이드를 이동하세요.
    </div>
    <button class="hint-ok" onclick="dismissHint()">확인</button>
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
      updateHeader();
      break;

    case 'position':
      if (typeof msg.idx === 'number') {
        currentIdx = msg.idx;
        updateHeader();
        highlightCurrent();
        scrollToActive();
      }
      break;

    case 'navigate':
      if (msg.direction === 'jump' && typeof msg.idx === 'number') {
        currentIdx = msg.idx;
        updateHeader();
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
  const dot = document.getElementById('ws-dot');
  dot.className = state;
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
  const btn = document.getElementById('btn-timer');
  const icon = document.getElementById('timer-icon');
  const label = document.getElementById('timer-label');
  if (timerEnabled) {
    btn.classList.add('active');
    icon.innerHTML = '&#9646;&#9646;';
    label.textContent = '타이머 ON';
  } else {
    btn.classList.remove('active');
    icon.innerHTML = '&#9654;';
    label.textContent = '타이머';
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
  setStreamBtnState('starting', '시작 중...');
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
    updateStreamBtn();
  })
  .catch(function() {
    streamBusy = false;
    updateStreamBtn();
    showToast('연결 오류');
  });
}

function doStopStream() {
  streamBusy = true;
  setStreamBtnState('starting', '종료 중...');
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
    updateStreamBtn();
  })
  .catch(function() {
    streamBusy = false;
    updateStreamBtn();
    showToast('연결 오류');
  });
}

function setStreamBtnState(cls, label) {
  const btn = document.getElementById('btn-stream');
  const dot = document.getElementById('stream-dot');
  const labelEl = document.getElementById('stream-label');
  btn.className = 'ctrl-btn ' + cls;
  dot.style.background = cls === 'live' ? '#ef5350' : (cls === 'starting' ? '#ffb74d' : '#555');
  labelEl.textContent = label;
}

function updateStreamBtn() {
  if (streamLive) {
    setStreamBtnState('live', 'LIVE');
  } else {
    setStreamBtnState('', 'LIVE');
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
function renderList() {
  const list = document.getElementById('order-list');
  const empty = document.getElementById('empty-msg');

  if (!items || items.length === 0) {
    list.innerHTML = '';
    empty.style.display = 'block';
    return;
  }
  empty.style.display = 'none';

  list.innerHTML = items.map(function(item, i) {
    const title = item.title || '';
    const obj = item.obj || '';
    const isActive = i === currentIdx;
    return '<div class="order-item' + (isActive ? ' active' : '') + '" onclick="jumpTo(' + i + ')" data-idx="' + i + '">' +
      '<span class="item-num">' + (i + 1) + '</span>' +
      '<span class="item-dot"></span>' +
      '<span class="item-title">' + escHtml(title) + '</span>' +
      (obj ? '<span class="item-obj">' + escHtml(obj) + '</span>' : '') +
      '</div>';
  }).join('');
}

function highlightCurrent() {
  const allItems = document.querySelectorAll('.order-item');
  allItems.forEach(function(el) {
    const idx = parseInt(el.getAttribute('data-idx'), 10);
    if (idx === currentIdx) {
      el.classList.add('active');
    } else {
      el.classList.remove('active');
    }
  });
}

function scrollToActive() {
  const active = document.querySelector('.order-item.active');
  if (active) {
    active.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }
}

function updateHeader() {
  const titleEl = document.getElementById('current-title');
  const progressEl = document.getElementById('progress-text');

  if (items && items.length > 0 && currentIdx >= 0 && currentIdx < items.length) {
    titleEl.textContent = items[currentIdx].title || '';
    progressEl.textContent = (currentIdx + 1) + '/' + items.length;
  } else {
    titleEl.textContent = '순서 없음';
    progressEl.textContent = '-/-';
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
        updateStreamBtn();
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
  // 캐시 버스팅으로 항상 최신 IP QR 표시
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

  // 수직 스크롤이 더 크면 리스트 스크롤로 처리
  if (Math.abs(dy) > Math.abs(dx)) return;

  // 수평 스와이프: 60px 이상
  if (Math.abs(dx) < 60) return;

  // 리스트 영역 터치면 스와이프 무시 (스크롤 허용)
  const wrap = document.getElementById('order-list-wrap');
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

  // 첫 방문 힌트 표시
  try {
    if (!localStorage.getItem('ep_remote_hint_seen')) {
      document.getElementById('swipe-hint').classList.add('show');
    }
  } catch(_) {}
})();
</script>
</body>
</html>`
