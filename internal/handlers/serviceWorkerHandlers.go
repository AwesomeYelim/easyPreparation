package handlers

import (
	"fmt"
	"net/http"
)

// HandleDisplaySW — GET /display/sw.js
// Display 페이지 오프라인 캐시용 Service Worker.
// 서버가 다운된 상태에서 프로젝터 화면이 새로고침되어도 캐시된 페이지를 서빙한다.
func HandleDisplaySW(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	// Service Worker 스코프를 /display 로 제한
	w.Header().Set("Service-Worker-Allowed", "/display")
	// SW 파일 자체는 캐시하지 않음 — 항상 최신 버전 확인
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	fmt.Fprint(w, `/* easyPreparation Display Service Worker v1 */
const CACHE = 'ep-display-v1';

/* 설치 시 Display 페이지 사전 캐시 */
self.addEventListener('install', function(e) {
  e.waitUntil(
    caches.open(CACHE).then(function(c) {
      return c.add('/display').catch(function(){});
    })
  );
  self.skipWaiting();
});

/* 구버전 캐시 정리 */
self.addEventListener('activate', function(e) {
  e.waitUntil(
    caches.keys().then(function(keys) {
      return Promise.all(
        keys.filter(function(k){ return k !== CACHE; }).map(function(k){ return caches.delete(k); })
      );
    }).then(function(){ return self.clients.claim(); })
  );
});

/* fetch 인터셉트 — 네트워크 우선, 실패 시 캐시 */
self.addEventListener('fetch', function(e) {
  var url = e.request.url;

  /* WebSocket, API, 비디오 파일은 SW에서 처리하지 않음 */
  if (url.indexOf('/ws') !== -1) return;
  if (url.indexOf('/api/') !== -1) return;
  if (url.indexOf('/display/video-bg/') !== -1) return;
  if (url.indexOf('/display/status') !== -1) return;

  e.respondWith(
    fetch(e.request).then(function(res) {
      /* Display 관련 페이지는 캐시 갱신 */
      if (res.ok && (
        url.indexOf('/display') !== -1 ||
        url.indexOf('/display/assets/') !== -1
      )) {
        var clone = res.clone();
        caches.open(CACHE).then(function(c){ c.put(e.request, clone); });
      }
      return res;
    }).catch(function() {
      /* 네트워크 실패 → 캐시에서 서빙 */
      return caches.match(e.request).then(function(cached) {
        if (cached) return cached;
        /* 캐시에도 없으면 /display 폴백 */
        return caches.match('/display');
      });
    })
  );
});
`)
}
