
// v2: /mobile 이 React 페이지로 바뀜 — 예전 HTML 오프라인 캐시를 지운다
const CACHE_NAME = 'ep-remote-v2';
const OFFLINE_URL = '/mobile/';

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
