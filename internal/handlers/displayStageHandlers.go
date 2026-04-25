package handlers

import "net/http"

// DisplayStageHandler — /display/stage
// 찬양팀·설교자용 무대 모니터 화면.
// 현재 슬라이드(크게) + 다음 항목 미리보기 + 경과 타이머
// WebSocket 'order' / 'navigate' / 'position' 메시지 재사용 (추가 서버 변경 없음)
func DisplayStageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(displayStageHTML))
}

const displayStageHTML = `<!DOCTYPE html>
<html lang="ko">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Stage Display</title>
<style>
  * { margin:0; padding:0; box-sizing:border-box; }

  html, body {
    width:100%; height:100%;
    background:#0d0d0d;
    color:#fff;
    font-family:'Malgun Gothic','맑은 고딕','Apple SD Gothic Neo',sans-serif;
    overflow:hidden;
    user-select:none;
  }

  /* ── 전체 레이아웃 ── */
  #root {
    display:grid;
    grid-template-rows: 1fr 64px;
    grid-template-columns: 1fr 300px;
    height:100vh;
    gap:0;
  }

  /* ── 현재 슬라이드 (좌측 메인) ── */
  #current {
    grid-row:1; grid-column:1;
    display:flex;
    flex-direction:column;
    justify-content:center;
    align-items:center;
    padding:48px 56px;
    border-right:1px solid rgba(255,255,255,0.08);
    border-bottom:1px solid rgba(255,255,255,0.08);
    position:relative;
    overflow:hidden;
  }

  #cur-label {
    position:absolute; top:20px; left:28px;
    font-size:13px; color:rgba(255,255,255,0.3);
    text-transform:uppercase; letter-spacing:0.12em;
    font-weight:600;
  }

  #cur-title {
    font-size:clamp(32px, 6vw, 72px);
    font-weight:700;
    text-align:center;
    line-height:1.25;
    color:#fff;
    margin-bottom:20px;
    text-shadow:0 2px 16px rgba(0,0,0,0.5);
    transition:opacity .25s;
  }

  #cur-obj {
    font-size:clamp(22px, 3.5vw, 46px);
    font-weight:400;
    text-align:center;
    color:rgba(255,255,255,0.75);
    line-height:1.5;
    transition:opacity .25s;
  }

  #cur-content {
    font-size:clamp(18px, 2.8vw, 38px);
    font-weight:500;
    text-align:center;
    color:rgba(255,255,255,0.9);
    line-height:1.8;
    white-space:pre-wrap;
    margin-top:18px;
    transition:opacity .25s;
  }

  #cur-page {
    position:absolute; bottom:20px; right:28px;
    font-size:13px; color:rgba(255,255,255,0.25);
  }

  /* ── 다음 항목 미리보기 (우측) ── */
  #next-panel {
    grid-row:1; grid-column:2;
    display:flex;
    flex-direction:column;
    padding:20px 16px;
    border-bottom:1px solid rgba(255,255,255,0.08);
    overflow:hidden;
  }

  #next-label {
    font-size:11px; color:rgba(255,255,255,0.3);
    text-transform:uppercase; letter-spacing:0.12em;
    font-weight:600; margin-bottom:14px;
  }

  #next-card {
    background:rgba(255,255,255,0.04);
    border:1px solid rgba(255,255,255,0.1);
    border-radius:12px;
    padding:18px 16px;
    flex:1;
    display:flex;
    flex-direction:column;
    justify-content:center;
    gap:8px;
    transition:all .3s;
  }

  #next-card.empty {
    justify-content:center;
    align-items:center;
  }

  #next-title {
    font-size:20px; font-weight:700; color:#fff;
    line-height:1.3;
  }

  #next-obj {
    font-size:15px; color:rgba(255,255,255,0.55);
    line-height:1.4;
  }

  #next-type {
    font-size:11px; color:rgba(59,130,246,0.8);
    font-weight:600; text-transform:uppercase;
    letter-spacing:0.1em; margin-bottom:4px;
  }

  #next-empty-msg {
    font-size:13px; color:rgba(255,255,255,0.2);
  }

  /* ── 하단 상태바 ── */
  #statusbar {
    grid-row:2; grid-column:1 / 3;
    display:flex;
    align-items:center;
    padding:0 28px;
    gap:24px;
    background:#111;
    border-top:1px solid rgba(255,255,255,0.08);
  }

  .stat {
    display:flex; align-items:center; gap:8px;
    color:rgba(255,255,255,0.5); font-size:14px;
  }
  .stat-icon {
    font-size:16px; opacity:.6;
  }
  .stat-val {
    font-variant-numeric:tabular-nums;
    font-weight:600; color:rgba(255,255,255,0.85);
  }

  #timer-val { color:#34d399; }
  #pos-val   { color:#60a5fa; }

  .stat-sep {
    flex:1;
  }

  #conn-dot {
    width:8px; height:8px; border-radius:50%;
    background:#ef4444;
    transition:background .3s;
  }
  #conn-dot.on { background:#34d399; }

  #cur-name-bar {
    font-size:13px; color:rgba(255,255,255,0.45);
    max-width:360px;
    overflow:hidden; text-overflow:ellipsis; white-space:nowrap;
  }

  /* ── 대기 화면 ── */
  #waiting {
    position:fixed; inset:0;
    background:#0d0d0d;
    display:flex; flex-direction:column;
    align-items:center; justify-content:center;
    gap:16px; z-index:100;
  }
  #waiting p { font-size:18px; color:rgba(255,255,255,0.4); }
  #waiting small { font-size:13px; color:rgba(255,255,255,0.2); }
  #waiting.hidden { display:none; }
</style>
</head>
<body>

<div id="waiting">
  <p>예배 순서를 기다리는 중...</p>
  <small>프로젝터에 보내기 후 자동으로 표시됩니다</small>
</div>

<div id="root">
  <!-- 현재 슬라이드 -->
  <div id="current">
    <div id="cur-label">현재</div>
    <div id="cur-title">—</div>
    <div id="cur-obj"></div>
    <div id="cur-content"></div>
    <div id="cur-page"></div>
  </div>

  <!-- 다음 항목 -->
  <div id="next-panel">
    <div id="next-label">다음</div>
    <div id="next-card">
      <div id="next-type"></div>
      <div id="next-title">—</div>
      <div id="next-obj"></div>
    </div>
  </div>

  <!-- 상태바 -->
  <div id="statusbar">
    <div class="stat">
      <span class="stat-icon">⏱</span>
      <span id="timer-val" class="stat-val">0:00</span>
    </div>
    <div class="stat">
      <span class="stat-icon">≡</span>
      <span id="pos-val" class="stat-val">— / —</span>
    </div>
    <div id="cur-name-bar"></div>
    <div class="stat-sep"></div>
    <div id="conn-dot"></div>
  </div>
</div>

<script>
var slides = [];
var idx = 0;
var subPageIdx = 0;
var subPages = [];
var ws = null;
var reconnectTimer = null;
var slotStart = Date.now();
var timerInterval = null;

/* ── 유틸 ── */
function sv(item, key) {
  if (!item) return '';
  var v = item[key];
  return (v != null) ? String(v) : '';
}

/* ── WS 연결 ── */
function connect() {
  var proto = location.protocol === 'https:' ? 'wss' : 'ws';
  ws = new WebSocket(proto + '://' + location.host + '/ws');

  ws.onopen = function() {
    document.getElementById('conn-dot').classList.add('on');
  };
  ws.onclose = function() {
    document.getElementById('conn-dot').classList.remove('on');
    clearTimeout(reconnectTimer);
    reconnectTimer = setTimeout(connect, 3000);
  };
  ws.onerror = function() {};

  ws.onmessage = function(e) {
    var msg;
    try { msg = JSON.parse(e.data); } catch(err) { return; }

    if (msg.type === 'order') {
      slides = msg.items || [];
      var start = (typeof msg.idx === 'number') ? msg.idx : 0;
      showSlide(start, 0);
      document.getElementById('waiting').classList.add('hidden');
    }

    if (msg.type === 'navigate') {
      if (msg.direction === 'jump' && typeof msg.idx === 'number') {
        showSlide(msg.idx, msg.subPageIdx || 0);
      } else if (msg.direction === 'jump_sub' && typeof msg.subPageIdx === 'number') {
        subPageIdx = msg.subPageIdx;
        renderCurrentContent();
      } else if (msg.direction === 'next') {
        advancePage(1);
      } else if (msg.direction === 'prev') {
        advancePage(-1);
      }
    }

    if (msg.type === 'position') {
      if (typeof msg.idx === 'number' && msg.idx !== idx) {
        showSlide(msg.idx, msg.subPageIdx || 0);
      }
    }
  };
}

/* ── 슬라이드 표시 ── */
function showSlide(i, sub) {
  if (!slides.length) return;
  i = Math.max(0, Math.min(i, slides.length - 1));
  idx = i;
  subPageIdx = sub || 0;
  subPages = buildSubPages(slides[idx]);

  slotStart = Date.now();
  renderCurrentContent();
  renderNext();
  updateStatusBar();
}

function advancePage(dir) {
  if (subPages.length > 1) {
    subPageIdx = Math.max(0, Math.min(subPageIdx + dir, subPages.length - 1));
  } else {
    var ni = idx + dir;
    if (ni >= 0 && ni < slides.length) showSlide(ni, 0);
    return;
  }
  renderCurrentContent();
}

/* ── 서브 페이지 빌드 ── */
function buildSubPages(item) {
  if (!item) return [];
  var info = sv(item, 'info');
  var title = sv(item, 'title');

  if (info.indexOf('b_') === 0 && item.contents) {
    return paginateText(item.contents, 3);
  }
  if (title === '신앙고백' && item.contents) {
    return paginateText(item.contents, 10);
  }
  if (info === 'lyrics_display' && item.pages && item.pages.length) {
    return item.pages;
  }
  if (title === '성시교독' && item.images && item.images.length) {
    return item.images.map(function(u) { return '[img]' + u; });
  }
  if (item.images && item.images.length) {
    return ['__cover__'].concat(item.images.map(function(u) { return '[img]' + u; }));
  }
  return [];
}

function paginateText(text, linesPerPage) {
  var lines = text.split('\n');
  var pages = [];
  for (var i = 0; i < lines.length; i += linesPerPage) {
    pages.push(lines.slice(i, i + linesPerPage).join('\n'));
  }
  return pages;
}

/* ── 현재 슬라이드 렌더 ── */
function renderCurrentContent() {
  var item = slides[idx];
  if (!item) return;

  var title   = sv(item, 'title');
  var obj     = sv(item, 'obj');
  var lead    = sv(item, 'lead');
  var info    = sv(item, 'info');
  var curPage = subPages.length > 1
    ? (subPageIdx + 1) + ' / ' + subPages.length
    : '';

  // 라벨
  document.getElementById('cur-label').textContent =
    lead ? lead + '  ·  현재' : '현재';

  // 제목
  document.getElementById('cur-title').textContent = title || '—';

  // obj / 내용
  var objEl = document.getElementById('cur-obj');
  var contentEl = document.getElementById('cur-content');

  objEl.textContent = '';
  contentEl.textContent = '';

  var sub = subPages.length > 0 ? subPages[subPageIdx] : null;

  if (sub && sub.indexOf('[img]') === 0) {
    // 이미지 페이지 — 이미지 번호 표시
    objEl.textContent = '이미지 ' + (subPageIdx + 1) + ' / ' + subPages.length;
  } else if (sub && sub !== '__cover__') {
    // 텍스트 페이지 (성경/가사/신앙고백 등)
    if (obj) objEl.textContent = obj;
    contentEl.textContent = sub;
  } else if (sub === '__cover__') {
    // 찬송 표지
    objEl.textContent = obj || '';
    contentEl.textContent = '';
  } else {
    // 일반 항목
    objEl.textContent = obj || '';
  }

  document.getElementById('cur-page').textContent = curPage;
  updateStatusBar();
}

/* ── 다음 항목 렌더 ── */
function renderNext() {
  var nextItem = slides[idx + 1] || null;
  var card = document.getElementById('next-card');
  var typeEl = document.getElementById('next-type');
  var titleEl = document.getElementById('next-title');
  var objEl = document.getElementById('next-obj');

  if (!nextItem) {
    card.innerHTML = '<span id="next-empty-msg">마지막 순서입니다</span>';
    return;
  }

  // 카드 초기화
  card.innerHTML =
    '<div id="next-type"></div>' +
    '<div id="next-title"></div>' +
    '<div id="next-obj"></div>';

  var title = sv(nextItem, 'title');
  var obj   = sv(nextItem, 'obj');
  var info  = sv(nextItem, 'info');

  // 타입 배지
  var typeLabel = '';
  if (info.indexOf('b_') === 0)       typeLabel = '성경';
  else if (info === 'lyrics_display') typeLabel = '찬양';
  else if (title === '찬송' || title === '헌금봉헌') typeLabel = '찬송';
  else if (title === '성시교독')       typeLabel = '교독';
  else if (title === '대표기도')       typeLabel = '기도';
  else if (title === '말씀')           typeLabel = '설교';

  document.getElementById('next-type').textContent  = typeLabel;
  document.getElementById('next-title').textContent = title || '—';
  document.getElementById('next-obj').textContent   = obj || '';
}

/* ── 상태바 업데이트 ── */
function updateStatusBar() {
  var item = slides[idx];
  document.getElementById('pos-val').textContent =
    slides.length ? (idx + 1) + ' / ' + slides.length : '— / —';
  document.getElementById('cur-name-bar').textContent =
    item ? sv(item, 'title') : '';
}

/* ── 경과 타이머 ── */
function startTimer() {
  clearInterval(timerInterval);
  timerInterval = setInterval(function() {
    var sec = Math.floor((Date.now() - slotStart) / 1000);
    var m = Math.floor(sec / 60);
    var s = sec % 60;
    document.getElementById('timer-val').textContent =
      m + ':' + (s < 10 ? '0' : '') + s;
  }, 1000);
}

connect();
startTimer();
</script>
</body>
</html>`
