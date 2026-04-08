package handlers

import (
	"easyPreparation_1.0/internal/assets"
	"easyPreparation_1.0/internal/obs"
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/quote"
	"easyPreparation_1.0/internal/utils"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

// 현재 예배 순서 메모리 저장
var (
	orderMu          sync.RWMutex
	currentOrder     []map[string]interface{}
	currentIdx       int
	displayChurchName string
)

// ── Display 상태 파일 영속화 ──

func displayStatePath() string {
	execPath := path.ExecutePath("easyPreparation")
	return filepath.Join(execPath, "data", "display_state.json")
}

// saveDisplayState — 현재 order+idx를 파일에 저장 (orderMu 잠긴 상태에서 호출하지 말 것)
func saveDisplayState() {
	orderMu.RLock()
	snapshot := deepCopyOrder(currentOrder)
	idx := currentIdx
	cn := displayChurchName
	orderMu.RUnlock()

	state := map[string]interface{}{
		"items":      snapshot,
		"idx":        idx,
		"churchName": cn,
	}

	data, err := json.Marshal(state)
	if err != nil {
		log.Printf("[display] 상태 저장 실패 (marshal): %v", err)
		return
	}
	if err := os.WriteFile(displayStatePath(), data, 0644); err != nil {
		log.Printf("[display] 상태 저장 실패 (write): %v", err)
	}
}

// deepCopyOrder — slice of maps를 완전한 깊은 복사 (JSON round-trip)
// orderMu 잠긴 상태에서 호출할 것 (잠금 없이 내부적으로 안전)
func deepCopyOrder(src []map[string]interface{}) []map[string]interface{} {
	if len(src) == 0 {
		return nil
	}
	data, err := json.Marshal(src)
	if err != nil {
		// fallback: 얕은 복사
		cp := make([]map[string]interface{}, len(src))
		copy(cp, src)
		return cp
	}
	var dst []map[string]interface{}
	if err := json.Unmarshal(data, &dst); err != nil {
		cp := make([]map[string]interface{}, len(src))
		copy(cp, src)
		return cp
	}
	return dst
}

// getOrderSnapshotLocked — orderMu 잠긴 상태에서 호출. 깊은 복사 + idx + churchName 반환.
func getOrderSnapshotLocked() ([]map[string]interface{}, int, string) {
	return deepCopyOrder(currentOrder), currentIdx, displayChurchName
}

// LoadDisplayState — 서버 시작 시 파일에서 복원
func LoadDisplayState() {
	data, err := os.ReadFile(displayStatePath())
	if err != nil {
		return // 파일 없으면 무시
	}
	var state struct {
		Items      []map[string]interface{} `json:"items"`
		Idx        int                      `json:"idx"`
		ChurchName string                   `json:"churchName"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		log.Printf("[display] 상태 복원 실패: %v", err)
		return
	}
	if len(state.Items) == 0 {
		return
	}
	orderMu.Lock()
	currentOrder = state.Items
	currentIdx = state.Idx
	displayChurchName = state.ChurchName
	orderMu.Unlock()
	log.Printf("[display] 상태 복원: %d개 항목, idx=%d", len(state.Items), state.Idx)
}

// ── 서버 사이드 자동 넘김 타이머 ──
var (
	timerMu          sync.Mutex
	timerEnabled     bool
	timerSpeedFactor float64 = 1.0
	timerCancel      chan struct{}
	timerCountdown   int
	timerCurIdx      int
	timerCurSubPage  int
)

const lordsPrayer = `하늘에 계신 우리 아버지,
아버지의 이름을 거룩하게 하시며, 아버지의 나라가 오게 하시며,
아버지의 뜻이 하늘에서와 같이 땅에서도 이루어지게 하소서.
오늘 우리에게 일용할 양식을 주시고,
우리가 우리에게 잘못한 사람을 용서하여 준 것같이 우리 죄를 용서하여 주시고,
우리를 시험에 빠지지 않게 하시고, 악에서 구하소서.
나라와 권능과 영광이 영원히 아버지의 것입니다. 아멘.`

const apostlesCreed = `나는 전능하신 아버지 하나님, 천지의 창조주를 믿습니다.
나는 그의 유일하신 아들, 우리 주 예수 그리스도를 믿습니다.
그는 성령으로 잉태되어 동정녀 마리아에게서 나시고,
본디오 빌라도에게 고난을 받아 십자가에 못 박혀 죽으시고,
장사된 지 사흘만에 죽은 자 가운데서 다시 살아나셨으며,
하늘에 오르시어 전능하신 아버지 하나님 우편에 앉아 계시다가,
거기로부터 살아있는 자와 죽은 자를 심판하러 오십니다.
나는 성령을 믿으며,
거룩한 공교회와 성도의 교제와 죄를 용서 받는 것과
몸의 부활과 영생을 믿습니다. 아멘.`

const displayHTML = `<!DOCTYPE html>
<html lang="ko">
<head>
<meta charset="UTF-8">
<title>Display</title>
<style>
  @font-face {
    font-family:'JacquesFrancois';
    src:url('/display/font/JacquesFrancois-regular.ttf') format('truetype');
    font-weight:400; font-style:normal;
  }
  * { margin:0; padding:0; box-sizing:border-box; }
  html, body {
    width:100%; height:100%;
    background:#000;
    color:#fff;
    font-family:'Malgun Gothic','맑은 고딕','Apple SD Gothic Neo',sans-serif;
    overflow:hidden;
    user-select:none;
  }

  #slide {
    width:100%; height:100vh;
    display:flex; flex-direction:column;
    justify-content:center; align-items:center;
    padding:7.4vh 6.25vw;
    position:relative;
    background-size:cover;
    background-position:center;
    background-repeat:no-repeat;
    opacity:0;
    transition:opacity 0.4s ease;
  }
  #slide.visible { opacity:1; }

  /* 좌상단 인도자 */
  .label {
    position:absolute; top:4.4vh; left:4.2vw;
    font-size:2.6vh; color:rgba(255,255,255,0.5);
    letter-spacing:0.05em;
  }
  /* 우상단 순서 제목 */
  .order-title {
    position:absolute; top:4.4vh; right:4.2vw;
    font-size:2.6vh; color:rgba(255,255,255,0.5);
  }
  /* 페이지 표시 */
  .page-indicator {
    position:absolute; bottom:3vh; right:4.2vw;
    font-size:2vh; color:rgba(255,255,255,0.35);
  }

  /* 기본 제목 */
  .title {
    font-size:5.9vh; font-weight:700;
    text-align:center; margin-bottom:3.3vh;
    line-height:1.3; letter-spacing:-0.01em;
    text-shadow:0 2px 8px rgba(0,0,0,0.6);
  }
  .obj {
    font-size:4.4vh; font-weight:400;
    color:rgba(255,255,255,0.85);
    text-align:center; line-height:1.5;
    text-shadow:0 2px 6px rgba(0,0,0,0.5);
  }

  /* 성경 본문 */
  .bible-ref {
    font-size:2.8vh; color:rgba(255,255,255,0.65);
    margin-bottom:3vh; text-align:left; width:100%;
    text-shadow:0 1px 4px rgba(0,0,0,0.5);
  }
  .bible-contents {
    font-size:3.3vh; line-height:2;
    text-align:left; color:#fff;
    white-space:pre-wrap; width:100%;
    text-shadow:0 1px 6px rgba(0,0,0,0.6);
  }

  /* 찬송/교독 큰 텍스트 */
  .hymn-number {
    font-size:10vh; font-weight:700;
    text-align:center; line-height:1.2;
    text-shadow:0 3px 12px rgba(0,0,0,0.7);
  }
  .hymn-sub {
    font-size:3.2vh; color:rgba(255,255,255,0.6);
    margin-top:2vh; text-align:center;
  }

  /* 이미지 슬라이드 (찬송/교독 스캔) */
  .slide-image {
    max-width:90vw; max-height:85vh;
    object-fit:contain;
  }

  /* 기도자 이름 대형 */
  .prayer-name {
    font-size:8vh; font-weight:700;
    text-align:center; line-height:1.3;
    text-shadow:0 3px 12px rgba(0,0,0,0.7);
  }

  /* 가사 슬라이드 (큰 텍스트 중앙) */
  .lyrics-text {
    font-size:5vh; line-height:1.8;
    text-align:center; color:#fff;
    white-space:pre-wrap; width:100%;
    font-weight:500;
    text-shadow:0 2px 10px rgba(0,0,0,0.7);
  }

  /* 신앙고백 (사도신경) 중앙정렬 */
  .creed-text {
    font-size:3.3vh; line-height:1.9;
    text-align:center; color:rgba(255,255,255,0.9);
    white-space:pre-wrap; width:100%;
    text-shadow:0 1px 6px rgba(0,0,0,0.5);
  }

  /* 참회의 기도 좌정렬 */
  .confession-text {
    font-size:3.6vh; line-height:2;
    text-align:left; color:rgba(255,255,255,0.9);
    white-space:pre-wrap; width:100%;
    text-shadow:0 1px 6px rgba(0,0,0,0.5);
  }

  /* 말씀 (설교) */
  .sermon-title {
    font-size:7vh; font-weight:700;
    text-align:center; line-height:1.3;
    margin-bottom:3vh;
    text-shadow:0 3px 12px rgba(0,0,0,0.7);
  }
  .sermon-pastor {
    font-size:4vh; color:rgba(255,255,255,0.7);
    text-align:center;
    text-shadow:0 2px 6px rgba(0,0,0,0.5);
  }

  /* 공지 (교회소식) */
  .notice-title {
    font-size:4.8vh; font-weight:700;
    margin-bottom:2.6vh; color:#fff;
    text-shadow:0 2px 8px rgba(0,0,0,0.6);
  }
  .notice-contents {
    font-size:2.8vh; color:rgba(255,255,255,0.85);
    text-align:left; line-height:1.8;
    white-space:pre-wrap; width:100%;
    text-shadow:0 1px 4px rgba(0,0,0,0.5);
  }

  /* 하단 구분선 */
  .divider {
    position:absolute; bottom:5.6vh; left:4.2vw; right:4.2vw;
    height:1px; background:rgba(255,255,255,0.15);
  }

  /* 슬라이드 인덱스 바 (좌하단) */
  .slide-pos {
    position:absolute; bottom:3vh; left:4.2vw;
    font-size:2vh; color:rgba(255,255,255,0.35);
  }

  /* 교회명 (우하단) */
  .church-box {
    position:absolute; right:0; bottom:0;
    background:rgba(60,55,50,0.55);
    padding:1.2vh 2vw;
    font-family:'JacquesFrancois',serif;
    font-size:3.2vh; color:#fff;
    font-weight:400;
    letter-spacing:0.1em;
  }

  /* 카운트다운 오버레이 */
  #countdown-overlay {
    position:fixed; top:0; left:0; width:100%; height:100%;
    background:rgba(0,0,0,0.85);
    display:none; flex-direction:column;
    justify-content:center; align-items:center;
    z-index:9999;
  }
  #countdown-overlay.visible { display:flex; }
  #countdown-label {
    font-size:6vh; font-weight:600;
    color:rgba(255,255,255,0.8);
    margin-bottom:3vh;
  }
  #countdown-time {
    font-size:15vh; font-weight:700;
    font-family:'SF Mono','Consolas','Courier New',monospace;
    color:#fff; letter-spacing:0.1em;
  }

</style>
</head>
<body>
<div id="slide"></div>
<div id="countdown-overlay">
  <div id="countdown-label"></div>
  <div id="countdown-time"></div>
</div>

<script>
const slide = document.getElementById('slide');
let ws, reconnectTimer;
let slides = [];
let idx = 0;
let subPages = [];   // 성경 본문 or 이미지 페이지
let subPageIdx = 0;
let churchName = '';

/* ───── WebSocket ───── */
function connect() {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws';
  ws = new WebSocket(proto + '://' + location.host + '/ws');
  ws.onopen = () => { console.log('[Display] WS connected'); };
  ws.onerror = (e) => { console.error('[Display] WS error', e); };
  ws.onmessage = (e) => {
    const msg = JSON.parse(e.data);
    console.log('[Display] WS msg:', msg.type, msg.items ? msg.items.length + ' items' : '');
    if (msg.type === 'order') { if (msg.churchName) churchName = msg.churchName; loadOrder(msg.items, msg.idx); }
    if (msg.type === 'navigate') {
      if (msg.direction === 'jump' && typeof msg.idx === 'number') {
        showSlide(msg.idx);
        if (typeof msg.subPageIdx === 'number' && msg.subPageIdx > 0) {
          subPageIdx = msg.subPageIdx;
          renderItem(slides[idx], subPageIdx);
          reportPosition();
        }
      } else if (msg.direction === 'jump_sub' && typeof msg.subPageIdx === 'number') {
        subPageIdx = msg.subPageIdx;
        renderItem(slides[idx], subPageIdx);
        reportPosition();
      } else {
        navigate(msg.direction);
      }
    }
    if (msg.type === 'display') renderSingle(msg);
    if (msg.type === 'schedule_countdown') {
      var overlay = document.getElementById('countdown-overlay');
      document.getElementById('countdown-label').textContent = msg.label;
      document.getElementById('countdown-time').textContent =
        String(msg.minutes).padStart(2,'0') + ':' + String(msg.seconds).padStart(2,'0');
      overlay.classList.add('visible');
    }
    if (msg.type === 'schedule_started') {
      document.getElementById('countdown-overlay').classList.remove('visible');
    }
  };
  ws.onclose = () => {
    clearTimeout(reconnectTimer);
    reconnectTimer = setTimeout(connect, 3000);
  };
}

/* ───── 순서 로드 ───── */
function loadOrder(items, startIdx) {
  var prevLen = slides.length;
  slides = items || [];
  var start = (typeof startIdx === 'number') ? startIdx : 0;
  // append: 이전에 슬라이드가 있고 startIdx가 현재 idx와 같으면 이동하지 않음
  if (prevLen > 0 && start === idx && slides.length > prevLen) {
    return;
  }
  idx = 0;
  lastReportedIdx = -1;
  showSlide(start);
}

/* ───── 슬라이드 표시 ───── */
let lastReportedIdx = -1;
function showSlide(i) {
  if (!slides.length) return;
  i = Math.max(0, Math.min(i, slides.length - 1));
  idx = i;
  subPageIdx = 0;
  // 항목 변경 시 서버에 위치 보고
  if (idx !== lastReportedIdx) {
    lastReportedIdx = idx;
    reportPosition();
  }
  const item = slides[idx];

  subPages = [];

  const itemTitle = item.title || '';

  // 성경 본문 → 텍스트 페이지 분할
  if ((item.info || '').startsWith('b_') && item.contents) {
    subPages = paginate(item.contents, 3);
  }
  // 신앙고백 본문 → 페이지 분할
  else if (itemTitle === '신앙고백' && item.contents) {
    subPages = paginate(item.contents, 10);
  }
  // 가사 슬라이드 (텍스트 페이지)
  else if ((item.info || '') === 'lyrics_display' && item.pages && item.pages.length > 0) {
    subPages = item.pages;
  }
  // 성시교독 → 이미지만 (표지 없음)
  else if (itemTitle === '성시교독' && item.images && item.images.length > 0) {
    subPages = item.images;
  }
  // 찬송/헌금봉헌 이미지 → 표지 + 이미지 페이지
  else if (item.images && item.images.length > 0) {
    subPages = ['__cover__'].concat(item.images);
  }

  renderItem(item, 0);
}

/* ───── 키보드 / 네비게이션 ───── */
function reportPosition() {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({type:'position', idx:idx, subPageIdx:subPageIdx}));
  }
}

function navigate(dir) {
  if (dir === 'next') {
    if (subPages.length > 1 && subPageIdx < subPages.length - 1) {
      subPageIdx++;
      renderItem(slides[idx], subPageIdx);
      reportPosition();
    } else {
      showSlide(idx + 1);
    }
  } else if (dir === 'prev') {
    if (subPages.length > 1 && subPageIdx > 0) {
      subPageIdx--;
      renderItem(slides[idx], subPageIdx);
      reportPosition();
    } else {
      showSlide(idx - 1);
    }
  }
}

document.addEventListener('keydown', (e) => {
  if (e.key === 'ArrowRight' || e.key === 'PageDown' || e.key === ' ')
    navigate('next');
  else if (e.key === 'ArrowLeft' || e.key === 'PageUp')
    navigate('prev');
});

/* ───── 렌더링 (title 기반 분기) ───── */
function renderItem(item, pageIdx) {
  console.log('[Display] renderItem idx=' + idx + ' pageIdx=' + pageIdx + ' info=' + (item.info||''));
  const info     = item.info     || '';
  const title    = item.title    || '';
  const obj      = item.obj      || '';
  const lead     = item.lead     || '';
  const contents = item.contents || '';
  const images   = item.images   || [];
  const bgImage  = item.bgImage  || '';

  // 항목별 배경 이미지 (있으면 사용, 없으면 기본)
  // 컨텐츠 항목은 어두운 오버레이로 가독성 확보 (default 케이스에서 이미지 전용 항목은 재설정)
  if (bgImage) {
    slide.style.backgroundImage = "linear-gradient(rgba(0,0,0,0.35),rgba(0,0,0,0.35)), url('" + bgImage + "')";
  } else {
    slide.style.backgroundImage = "linear-gradient(rgba(0,0,0,0.4),rgba(0,0,0,0.4)), url('/display/bg')";
  }
  slide.className = 'visible';

  const posText = slides.length ? (idx + 1) + ' / ' + slides.length : '';
  const pageText = subPages.length > 1 ? (subPageIdx + 1) + ' / ' + subPages.length : '';
  const churchBox = (bgImage && churchName) ? '<div class="church-box">' + esc(churchName) + '</div>' : '';
  const footer =
    '<div class="divider"></div>' +
    '<div class="slide-pos">' + posText + '</div>' +
    (pageText ? '<div class="page-indicator">' + pageText + '</div>' : '') +
    churchBox;
  const header =
    '<div class="label">' + esc(lead) + '</div>' +
    '<div class="order-title">' + esc(title) + '</div>';

  // ── 1. 성경 본문 (b_edit + contents) ──
  if (info.startsWith('b_') && contents) {
    const page = subPages[pageIdx] || contents;
    slide.innerHTML = header +
      '<div class="bible-ref">' + esc(obj) + '</div>' +
      '<div class="bible-contents">' + esc(page) + '</div>' +
      footer;
    return;
  }

  // ── 1b. 가사 슬라이드 (lyrics_display) ──
  if (info === 'lyrics_display' && subPages.length > 0) {
    var lyricsPage = subPages[pageIdx] || '';
    console.log('[Display] lyrics render pageIdx=' + pageIdx + '/' + subPages.length + ' text=' + lyricsPage.substring(0, 30));
    slide.innerHTML = header +
      '<div class="lyrics-text">' + esc(lyricsPage) + '</div>' +
      footer;
    return;
  }

  // ── 2. 찬송 / 헌금봉헌 (표지 + 이미지 페이지) ──
  if (title === '찬송' || title === '헌금봉헌') {
    if (images.length > 0 && pageIdx > 0) {
      slide.style.backgroundImage = 'none';
      slide.innerHTML =
        '<img class="slide-image" src="' + images[pageIdx - 1] + '">' +
        footer;
      return;
    }
    slide.innerHTML = header +
      '<div class="hymn-number">' + esc(obj) + '</div>' +
      (lead ? '<div class="hymn-sub">' + esc(lead) + '</div>' : '') +
      footer;
    return;
  }

  // ── 3. 성시교독 (이미지만 — 표지 없이 바로 이미지) ──
  if (title === '성시교독') {
    if (images.length > 0) {
      slide.style.backgroundImage = 'none';
      slide.innerHTML =
        '<img class="slide-image" src="' + images[Math.min(pageIdx, images.length - 1)] + '">' +
        footer;
      return;
    }
    slide.innerHTML = header +
      '<div class="hymn-number">' + esc(obj) + '</div>' +
      (lead ? '<div class="hymn-sub">' + esc(lead) + '</div>' : '') +
      footer;
    return;
  }

  // ── 4. 대표기도 ──
  if (title === '대표기도') {
    slide.innerHTML = header +
      '<div class="title">' + esc(title) + '</div>' +
      '<div class="prayer-name">' + esc(lead) + '</div>' +
      footer;
    return;
  }

  // ── 5. 신앙고백 (사도신경 등) ──
  if (title === '신앙고백' && contents) {
    const page = subPages.length > 0 ? (subPages[pageIdx] || contents) : contents;
    slide.innerHTML = header +
      '<div class="title">' + esc(obj) + '</div>' +
      '<div class="creed-text">' + esc(page) + '</div>' +
      footer;
    return;
  }

  // ── 5b. 주기도문 ──
  if (title === '주기도문' && contents) {
    const page = subPages.length > 0 ? (subPages[pageIdx] || contents) : contents;
    slide.innerHTML = header +
      '<div class="title">' + esc(title) + '</div>' +
      '<div class="creed-text">' + esc(page) + '</div>' +
      footer;
    return;
  }

  // ── 6. 참회의 기도 (좌정렬 멀티라인, bgImage 있으면 제목 생략) ──
  if (title === '참회의 기도') {
    slide.innerHTML = header +
      (bgImage ? '' : '<div class="title">' + esc(title) + '</div>') +
      '<div class="confession-text">' + esc(obj !== '-' ? obj : '') + '</div>' +
      footer;
    return;
  }

  // ── 6. 말씀 (설교 제목 + 목사) ──
  if (title === '말씀') {
    slide.innerHTML = header +
      '<div class="sermon-title">' + esc(obj !== '-' ? obj : '') + '</div>' +
      (lead ? '<div class="sermon-pastor">' + esc(lead) + '</div>' : '') +
      footer;
    return;
  }

  // ── 7. 교회소식 (계층 텍스트) ──
  if (info === 'notice' || title === '교회소식') {
    slide.innerHTML = header +
      '<div class="notice-title">' + esc(title) + '</div>' +
      '<div class="notice-contents">' + esc(contents || obj) + '</div>' +
      footer;
    return;
  }

  // ── 8. 기본 (전주, 예배의 부름, 축도 등) ──
  // 배경 이미지가 있는 단순 항목은 이미지만 표시 (제목 텍스트가 이미지에 포함)
  if (bgImage) {
    slide.style.backgroundImage = "url('" + bgImage + "')";
    slide.innerHTML = churchBox;
    return;
  }
  var mainText = (obj && obj !== '-') ? obj : '';
  slide.innerHTML = header +
    '<div class="title">' + esc(title) + '</div>' +
    (mainText ? '<div class="obj">' + esc(mainText) + '</div>' : '') +
    (lead && !mainText ? '<div class="prayer-name">' + esc(lead) + '</div>' : '') +
    footer;
}

/* 단독 push 호환 */
function renderSingle(data) {
  slides = [data];
  idx = 0;
  showSlide(0);
}

/* ───── 유틸 ───── */
function paginate(text, linesPerPage) {
  const lines = text.split('\n').filter(l => l.trim() !== '');
  const pages = [];
  for (let i = 0; i < lines.length; i += linesPerPage) {
    pages.push(lines.slice(i, i + linesPerPage).join('\n'));
  }
  return pages.length ? pages : [text];
}

function esc(s) {
  return String(s)
    .replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;')
    .replace(/\n/g,'<br>');
}

connect();
</script>
</body>
</html>`

// UpdateDisplayIdx — display HTML이 WS로 보고한 현재 위치 업데이트
func UpdateDisplayIdx(newIdx int) {
	orderMu.Lock()
	defer orderMu.Unlock()
	if newIdx >= 0 && newIdx < len(currentOrder) {
		currentIdx = newIdx
		go saveDisplayState()
	}
}

// GetCurrentTitle — OBS 씬 전환용 현재 항목 title 조회
func GetCurrentTitle() string {
	orderMu.RLock()
	defer orderMu.RUnlock()
	if currentIdx >= 0 && currentIdx < len(currentOrder) {
		t, _ := currentOrder[currentIdx]["title"].(string)
		return t
	}
	return ""
}

// GetCurrentInfo — 현재 항목의 info 필드 조회
func GetCurrentInfo() string {
	orderMu.RLock()
	defer orderMu.RUnlock()
	if currentIdx >= 0 && currentIdx < len(currentOrder) {
		info, _ := currentOrder[currentIdx]["info"].(string)
		return info
	}
	return ""
}

// IsFadeBackItem — fade-back 대상 항목 여부 (현재 사용 안 함)
func IsFadeBackItem(title string) bool {
	return false
}

// ── 서버 사이드 타이머 함수 ──

// calcSlideDelayLocked — 현재 항목/서브페이지에 대한 딜레이(초) 계산
// 호출 시점에 timerMu가 이미 잠겨 있어야 하고, orderMu는 잠기지 않은 상태여야 한다.
// timerCurIdx, timerCurSubPage를 인자로 받아 lock re-entry를 방지.
func calcSlideDelayLocked(curIdx, curSubPage int) int {
	orderMu.RLock()
	defer orderMu.RUnlock()

	if curIdx < 0 || curIdx >= len(currentOrder) {
		return 0
	}
	item := currentOrder[curIdx]
	title, _ := item["title"].(string)
	info, _ := item["info"].(string)

	bpm := 0
	switch v := item["bpm"].(type) {
	case float64:
		bpm = int(v)
	case int:
		bpm = v
	}

	// 전주/후주: 60초
	if title == "전주" || title == "후주" {
		return 60
	}

	// 성시교독: 표지 없음, 이미지만 15초
	if title == "성시교독" {
		return 15
	}

	// 찬송/헌금봉헌 이미지: 커버 5초, 나머지 15초
	if title == "찬송" || title == "헌금봉헌" {
		hasImages := false
		if imgs, ok := item["images"].([]string); ok && len(imgs) > 0 {
			hasImages = true
		} else if imgs, ok := item["images"].([]interface{}); ok && len(imgs) > 0 {
			hasImages = true
		}
		if hasImages {
			if curSubPage == 0 {
				return 5
			}
			return 15
		}
	}

	// 가사 (lyrics_display): 글자 수 비례 + BPM 보정
	if info == "lyrics_display" {
		baseTime := 8.0
		if bpm > 0 {
			baseTime = float64(16) / float64(bpm) * 60
		}

		// pages에서 현재 슬라이드와 전체 평균 글자 수 계산
		pages := extractPages(item)
		if len(pages) > 0 && curSubPage >= 0 && curSubPage < len(pages) {
			avgChars := 0.0
			for _, p := range pages {
				avgChars += float64(countKoreanChars(p))
			}
			avgChars /= float64(len(pages))
			if avgChars > 0 {
				curChars := float64(countKoreanChars(pages[curSubPage]))
				ratio := curChars / avgChars
				// 비율 범위 제한 (0.5 ~ 2.0)
				if ratio < 0.5 {
					ratio = 0.5
				} else if ratio > 2.0 {
					ratio = 2.0
				}
				return int(math.Round(baseTime * ratio))
			}
		}
		return int(math.Round(baseTime))
	}

	// 그 외 (대표기도, 말씀, 교회소식 등): 수동
	return 0
}

// extractPages — item에서 pages 배열 추출
func extractPages(item map[string]interface{}) []string {
	switch v := item["pages"].(type) {
	case []string:
		return v
	case []interface{}:
		pages := make([]string, 0, len(v))
		for _, p := range v {
			if s, ok := p.(string); ok {
				pages = append(pages, s)
			}
		}
		return pages
	}
	return nil
}

// countKoreanChars — 공백/줄바꿈 제외한 글자 수 (한글+영문+숫자)
func countKoreanChars(s string) int {
	count := 0
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			count++
		}
	}
	return count
}

// restartServerTimer — 타이머 재시작 (timerMu 외부에서 호출)
func restartServerTimer() {
	timerMu.Lock()

	// 기존 타이머 취소
	if timerCancel != nil {
		close(timerCancel)
	}
	cancel := make(chan struct{})
	timerCancel = cancel

	if !timerEnabled {
		timerCountdown = 0
		timerMu.Unlock()
		broadcastTimerState()
		return
	}

	// timerMu를 잠근 상태에서 timer 변수를 로컬로 복사한 뒤 Unlock → calcSlideDelayLocked 호출
	// (calcSlideDelayLocked 내부에서 orderMu.RLock을 사용하므로 timerMu를 먼저 풀어야 교착 방지)
	curIdx := timerCurIdx
	curSubPage := timerCurSubPage
	speedFactor := timerSpeedFactor
	timerMu.Unlock()

	delay := calcSlideDelayLocked(curIdx, curSubPage)
	if delay <= 0 {
		timerMu.Lock()
		timerCountdown = 0
		timerMu.Unlock()
		broadcastTimerState()
		return
	}

	// 속도 적용
	delay = int(math.Round(float64(delay) / speedFactor))
	if delay < 1 {
		delay = 1
	}

	timerMu.Lock()
	timerCountdown = delay

	// goroutine 시작 후 Unlock — 빠른 호출 시 다중 goroutine 누수 방지
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-cancel:
				return
			case <-ticker.C:
				timerMu.Lock()
				timerCountdown--
				if timerCountdown <= 0 {
					timerCountdown = 0
					curI := timerCurIdx
					curS := timerCurSubPage
					timerMu.Unlock()
					broadcastTimerState()
					log.Printf("[timer] auto-navigate next (idx=%d subPage=%d)", curI, curS)
					BroadcastMessage("navigate", map[string]interface{}{"direction": "next"})
					return
				}
				timerMu.Unlock()
				broadcastTimerState()
			}
		}
	}()

	timerMu.Unlock()
	broadcastTimerState()
}

// stopServerTimer — 타이머 정지
func stopServerTimer() {
	timerMu.Lock()
	timerEnabled = false
	if timerCancel != nil {
		close(timerCancel)
		timerCancel = nil
	}
	timerCountdown = 0
	timerMu.Unlock()
	broadcastTimerState()
}

// broadcastTimerState — 제어판에 타이머 상태 전송
func broadcastTimerState() {
	timerMu.Lock()
	state := map[string]interface{}{
		"enabled":     timerEnabled,
		"countdown":   timerCountdown,
		"idx":         timerCurIdx,
		"subPageIdx":  timerCurSubPage,
		"speedFactor": timerSpeedFactor,
	}
	timerMu.Unlock()
	BroadcastMessage("timer_state", state)
}

// OnPositionUpdate — Display에서 위치 보고 시 호출 (websocket.go에서 호출)
func OnPositionUpdate(newIdx, newSubPage int) {
	timerMu.Lock()
	changed := newIdx != timerCurIdx || newSubPage != timerCurSubPage
	timerCurIdx = newIdx
	timerCurSubPage = newSubPage
	enabled := timerEnabled
	timerMu.Unlock()

	if enabled && changed {
		restartServerTimer()
	}
}

// ── /display/overlay — OBS 방송용 텍스트 오버레이 ──

const displayOverlayHTML = `<!DOCTYPE html>
<html lang="ko">
<head>
<meta charset="UTF-8">
<title>Display Overlay</title>
<style>
  :root {
    --overlay-font-size: 42px;
    --overlay-line-height: 1.7;
    --overlay-font-weight: 600;
    --overlay-color: #fff;
    --overlay-text-shadow: 0 2px 12px rgba(0,0,0,0.9), 0 0 4px rgba(0,0,0,0.7);
    --overlay-position: flex-end;
    --overlay-padding: 40px 60px;
    --overlay-bg: rgba(0,0,0,0.75);
    --overlay-bg-radius: 16px;
    --overlay-bg-padding: 28px 40px;
    --title-font-size: 48px;
    --title-font-weight: 700;
    --sub-font-size: 32px;
    --sub-color: rgba(255,255,255,0.8);
    --ref-font-size: 24px;
    --ref-color: rgba(255,255,255,0.7);
    --bible-font-size: 34px;
    --bible-line-height: 1.8;
    --transition-speed: 0.4s;
  }

  * { margin:0; padding:0; box-sizing:border-box; }
  html, body {
    width:100%; height:100%;
    background:transparent;
    color:var(--overlay-color);
    font-family:'Malgun Gothic','맑은 고딕','Apple SD Gothic Neo',sans-serif;
    overflow:hidden;
    user-select:none;
  }

  #slide {
    width:100%; height:100vh;
    display:flex; flex-direction:column;
    justify-content:var(--overlay-position);
    align-items:center;
    padding:var(--overlay-padding);
    position:relative;
    opacity:0;
    transition:opacity var(--transition-speed) ease;
  }
  #slide.visible { opacity:1; }

  .overlay-box {
    background:var(--overlay-bg);
    border-radius:var(--overlay-bg-radius);
    padding:var(--overlay-bg-padding);
    display:flex; flex-direction:column;
    align-items:flex-start;
    width:1500px;
  }
  .overlay-box.center { align-items:center; }

  .lyrics-overlay {
    font-size:var(--overlay-font-size);
    line-height:var(--overlay-line-height);
    text-align:center;
    width:100%;
    color:var(--overlay-color);
    white-space:pre-wrap;
    font-weight:var(--overlay-font-weight);
    text-shadow:var(--overlay-text-shadow);
  }

  .bible-overlay-ref {
    font-size:var(--ref-font-size);
    color:var(--ref-color);
    margin-bottom:2vh; text-align:left;
    text-shadow:var(--overlay-text-shadow);
  }
  .bible-overlay-text {
    font-size:var(--bible-font-size);
    line-height:var(--bible-line-height);
    text-align:left;
    color:var(--overlay-color);
    white-space:pre-wrap;
    text-shadow:var(--overlay-text-shadow);
  }

  .title-overlay {
    font-size:var(--title-font-size);
    font-weight:var(--title-font-weight);
    text-align:left;
    text-shadow:var(--overlay-text-shadow);
  }
  .sub-overlay {
    font-size:var(--sub-font-size);
    color:var(--sub-color);
    margin-top:1.5vh; text-align:left;
    text-shadow:var(--overlay-text-shadow);
  }

  /* 카운트다운 오버레이 */
  #countdown-overlay {
    position:fixed; top:0; left:0; width:100%; height:100%;
    background:rgba(0,0,0,0.85);
    display:none; flex-direction:column;
    justify-content:center; align-items:center;
    z-index:9999;
  }
  #countdown-overlay.visible { display:flex; }
  #countdown-label {
    font-size:6vh; font-weight:600;
    color:rgba(255,255,255,0.8);
    margin-bottom:3vh;
  }
  #countdown-time {
    font-size:15vh; font-weight:700;
    font-family:'SF Mono','Consolas','Courier New',monospace;
    color:#fff; letter-spacing:0.1em;
  }
</style>
</head>
<body>
<div id="slide"></div>
<div id="countdown-overlay">
  <div id="countdown-label"></div>
  <div id="countdown-time"></div>
</div>

<script>
const slide = document.getElementById('slide');
let ws, reconnectTimer;
let slides = [];
let idx = 0;
let subPages = [];
let subPageIdx = 0;

function connect() {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws';
  ws = new WebSocket(proto + '://' + location.host + '/ws');
  ws.onopen = () => { console.log('[Lyrics] WS connected'); };
  ws.onerror = (e) => { console.error('[Lyrics] WS error', e); };
  ws.onmessage = (e) => {
    const msg = JSON.parse(e.data);
    if (msg.type === 'order') loadOrder(msg.items, msg.idx);
    if (msg.type === 'navigate') {
      if (msg.direction === 'jump' && typeof msg.idx === 'number') {
        showSlide(msg.idx);
        if (typeof msg.subPageIdx === 'number' && msg.subPageIdx > 0) {
          subPageIdx = msg.subPageIdx;
          renderLyricsItem(slides[idx], subPageIdx);
        }
      } else if (msg.direction === 'jump_sub' && typeof msg.subPageIdx === 'number') {
        subPageIdx = msg.subPageIdx;
        renderLyricsItem(slides[idx], subPageIdx);
      } else {
        navigate(msg.direction);
      }
    }
    if (msg.type === 'schedule_countdown') {
      var overlay = document.getElementById('countdown-overlay');
      document.getElementById('countdown-label').textContent = msg.label;
      document.getElementById('countdown-time').textContent =
        String(msg.minutes).padStart(2,'0') + ':' + String(msg.seconds).padStart(2,'0');
      overlay.classList.add('visible');
    }
    if (msg.type === 'schedule_started') {
      document.getElementById('countdown-overlay').classList.remove('visible');
    }
  };
  ws.onclose = () => {
    clearTimeout(reconnectTimer);
    reconnectTimer = setTimeout(connect, 3000);
  };
}

function loadOrder(items, startIdx) {
  var prevLen = slides.length;
  slides = items || [];
  var start = (typeof startIdx === 'number') ? startIdx : 0;
  if (prevLen > 0 && start === idx && slides.length > prevLen) return;
  idx = 0;
  showSlide(start);
}

function showSlide(i) {
  if (!slides.length) return;
  i = Math.max(0, Math.min(i, slides.length - 1));
  idx = i;
  subPageIdx = 0;
  const item = slides[idx];
  subPages = [];

  const itemTitle = item.title || '';

  if ((item.info || '').startsWith('b_') && item.contents) {
    subPages = paginate(item.contents, 3);
  } else if (itemTitle === '신앙고백' && item.contents) {
    subPages = paginate(item.contents, 10);
  } else if ((item.info || '') === 'lyrics_display' && item.pages && item.pages.length > 0) {
    subPages = item.pages;
  } else if (itemTitle === '성시교독' && item.images && item.images.length > 0) {
    subPages = item.images;
  } else if (item.images && item.images.length > 0) {
    subPages = ['__cover__'].concat(item.images);
  }

  renderLyricsItem(item, 0);
}

function navigate(dir) {
  if (dir === 'next') {
    if (subPages.length > 1 && subPageIdx < subPages.length - 1) {
      subPageIdx++;
      renderLyricsItem(slides[idx], subPageIdx);
    } else {
      showSlide(idx + 1);
    }
  } else if (dir === 'prev') {
    if (subPages.length > 1 && subPageIdx > 0) {
      subPageIdx--;
      renderLyricsItem(slides[idx], subPageIdx);
    } else {
      showSlide(idx - 1);
    }
  }
}

function renderLyricsItem(item, pageIdx) {
  const info     = item.info     || '';
  const title    = item.title    || '';
  const obj      = item.obj      || '';
  const contents = item.contents || '';
  const lead     = item.lead     || '';
  const bibleRef = item.bibleRef || '';
  const lyricsMap = item.lyricsMap || [];

  slide.className = 'visible';

  // 0. 말씀
  if (title === '말씀') {
    var sermonTitle = (obj && obj !== '-') ? obj : '';
    slide.innerHTML = '<div class="overlay-box center">' +
      '<div class="title-overlay" style="text-align:center;width:100%">' + esc(sermonTitle || title) + '</div>' +
      (lead ? '<div class="sub-overlay" style="text-align:center;width:100%">' + esc(lead) + '</div>' : '') +
      (bibleRef ? '<div class="sub-overlay" style="text-align:center;width:100%;margin-top:1vh;font-size:2.8vh;color:rgba(255,255,255,0.6);">' + esc(bibleRef) + '</div>' : '') +
      '</div>';
    return;
  }

  // 0b. 대표기도
  if (title === '대표기도') {
    slide.innerHTML = '<div class="overlay-box center">' +
      '<div class="title-overlay" style="text-align:center;width:100%">' + esc(title) + '</div>' +
      (lead ? '<div class="sub-overlay" style="text-align:center;width:100%;font-size:var(--title-font-size);font-weight:700;margin-top:2vh;">' + esc(lead) + '</div>' : '') +
      '</div>';
    return;
  }

  // 1a. 성시교독 — 오버레이 표시 없음 (이미지만 프로젝터에 표시)
  if (title === '성시교독') {
    slide.innerHTML = '';
    return;
  }

  // 1b. 찬양 — 할렐루야 성가대 표시
  if (title === '찬양') {
    var songTitle = (obj && obj !== '-') ? obj : '';
    slide.innerHTML = '<div class="overlay-box center">' +
      '<div class="title-overlay" style="text-align:center;width:100%">할렐루야 성가대</div>' +
      (songTitle ? '<div class="sub-overlay" style="text-align:center;width:100%;font-size:var(--sub-font-size);margin-top:2vh;">' + esc(songTitle) + '</div>' : '') +
      '</div>';
    return;
  }

  // 1. 찬송/헌금봉헌
  if (title === '찬송' || title === '헌금봉헌') {
    if (pageIdx === 0) {
      slide.innerHTML = '<div class="overlay-box center"><div class="title-overlay" style="text-align:center;width:100%">' + esc(obj) + '</div></div>';
      return;
    }
    var lyric = (pageIdx - 1 < lyricsMap.length) ? lyricsMap[pageIdx - 1] : '';
    slide.innerHTML = '<div class="overlay-box center"><div class="lyrics-overlay">' + esc(lyric) + '</div></div>';
    return;
  }

  // 2. 가사 슬라이드 (lyrics_display)
  if (info === 'lyrics_display' && subPages.length > 0) {
    var lyricsPage = subPages[pageIdx] || '';
    slide.innerHTML = '<div class="overlay-box center"><div class="lyrics-overlay">' + esc(lyricsPage) + '</div></div>';
    return;
  }

  // 3. 성경 본문
  if (info.startsWith('b_') && contents) {
    var page = subPages[pageIdx] || contents;
    slide.innerHTML = '<div class="overlay-box">' +
      '<div class="bible-overlay-ref">' + esc(obj) + '</div>' +
      '<div class="bible-overlay-text">' + esc(page) + '</div>' +
      '</div>';
    return;
  }

  // 4. 신앙고백/주기도문
  if ((title === '신앙고백' || title === '주기도문') && contents) {
    var creedPage = subPages.length > 0 ? (subPages[pageIdx] || contents) : contents;
    slide.innerHTML = '<div class="overlay-box center">' +
      '<div class="title-overlay" style="text-align:center;width:100%">' + esc(title) + '</div>' +
      '<div class="sub-overlay" style="white-space:pre-wrap;margin-top:3vh;text-align:center;width:100%">' + esc(creedPage) + '</div>' +
      '</div>';
    return;
  }

  // 5. 기타
  var subText = (obj && obj !== '-') ? obj : '';
  slide.innerHTML = '<div class="overlay-box center">' +
    '<div class="title-overlay" style="text-align:center;width:100%">' + esc(title) + '</div>' +
    (subText ? '<div class="sub-overlay" style="text-align:center;width:100%">' + esc(subText) + '</div>' : '') +
    '</div>';
}

function paginate(text, linesPerPage) {
  const lines = text.split('\n').filter(l => l.trim() !== '');
  const pages = [];
  for (let i = 0; i < lines.length; i += linesPerPage) {
    pages.push(lines.slice(i, i + linesPerPage).join('\n'));
  }
  return pages.length ? pages : [text];
}

function esc(s) {
  return String(s)
    .replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;')
    .replace(/\n/g,'<br>');
}

connect();
</script>
</body>
</html>`

// DisplayOverlayHandler — GET /display/overlay
func DisplayOverlayHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	_, _ = w.Write([]byte(displayOverlayHTML))
}

// DisplayHandler — GET /display
func DisplayHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	_, _ = w.Write([]byte(displayHTML))
}

// DisplayAssetsHandler — GET /display/assets/{name}.png
// 배경 이미지 서빙
func DisplayAssetsHandler(w http.ResponseWriter, r *http.Request) {
	name := filepath.Base(r.URL.Path)
	execPath := path.ExecutePath("easyPreparation")
	imgPath := filepath.Join(execPath, "data", "templates", "display", name)
	http.ServeFile(w, r, imgPath)
}

// DisplayBgHandler — GET /display/bg
// 공통 배경 이미지 서빙
func DisplayBgHandler(w http.ResponseWriter, r *http.Request) {
	execPath := path.ExecutePath("easyPreparation")
	imgPath := filepath.Join(execPath, "data", "templates", "lyrics", "Frame 2.png")
	http.ServeFile(w, r, imgPath)
}

// DisplayFontHandler — GET /display/font/{name}
func DisplayFontHandler(w http.ResponseWriter, r *http.Request) {
	name := filepath.Base(r.URL.Path)
	execPath := path.ExecutePath("easyPreparation")
	fontPath := filepath.Join(execPath, "public", "font", name)
	http.ServeFile(w, r, fontPath)
}

// DisplayTmpHandler — GET /display/tmp/{name}
// 찬송/교독 변환 PNG 서빙
func DisplayTmpHandler(w http.ResponseWriter, r *http.Request) {
	name := filepath.Base(r.URL.Path)
	execPath := path.ExecutePath("easyPreparation")
	imgPath := filepath.Join(execPath, "data", "cache", "hymn_pages", name)
	http.ServeFile(w, r, imgPath)
}

// DisplayOrderHandler — POST /display/order
// 전체 예배 순서 수신 → 성경 본문 자동 조회 → WebSocket broadcast
func DisplayOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var order []map[string]interface{}
	var raw json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	// wrapper format: {"items": [...], "churchName": "..."} 또는 plain array [...]
	var wrapper struct {
		Items        []map[string]interface{} `json:"items"`
		ChurchName   string                   `json:"churchName"`
		Email        string                   `json:"email"`
		Preprocessed bool                     `json:"preprocessed"`
	}
	var displayEmail string
	var skipPreprocess bool
	var newChurchName string
	if err := json.Unmarshal(raw, &wrapper); err == nil && len(wrapper.Items) > 0 {
		order = wrapper.Items
		newChurchName = wrapper.ChurchName
		displayEmail = wrapper.Email
		skipPreprocess = wrapper.Preprocessed
	} else if err := json.Unmarshal(raw, &order); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	BroadcastMessage("display_loading", map[string]interface{}{
		"message": "예배 화면 준비 중...",
		"total":   len(order),
	})

	// 항목별 전처리 (이미 전처리된 이력 데이터는 스킵)
	if !skipPreprocess {
		for i, item := range order {
			title, _ := item["title"].(string)
			BroadcastMessage("display_loading", map[string]interface{}{
				"message": fmt.Sprintf("%s 처리 중... (%d/%d)", title, i+1, len(order)),
				"current": i + 1,
				"total":   len(order),
			})
			order[i] = preprocessItem(item)
		}
	}

	// 말씀 항목에 직전 성경봉독의 구절 참조를 bibleRef로 주입
	var lastBibleRef string
	for i := range order {
		title, _ := order[i]["title"].(string)
		info, _ := order[i]["info"].(string)
		if strings.HasPrefix(info, "b_") && title != "말씀" {
			if ref, ok := order[i]["obj"].(string); ok && ref != "" && ref != "-" {
				lastBibleRef = ref
			}
		}
		if title == "말씀" && lastBibleRef != "" {
			order[i]["bibleRef"] = lastBibleRef
		}
	}

	// 기존 order에서 lyrics/bible source 항목 보존 (깊은 복사로 보존)
	orderMu.RLock()
	var preserved []map[string]interface{}
	for _, item := range currentOrder {
		if src, ok := item["source"].(string); ok && (src == "lyrics" || src == "bible") {
			preserved = append(preserved, item)
		}
	}
	orderMu.RUnlock()

	if len(preserved) > 0 {
		order = append(order, preserved...)
	}

	// 새 순서 로드 시 타이머 초기화
	stopServerTimer()

	orderMu.Lock()
	currentOrder = order
	currentIdx = 0
	displayChurchName = newChurchName
	cn := displayChurchName
	orderMu.Unlock()

	BroadcastMessage("display_loading", map[string]interface{}{
		"message": "준비 완료!",
		"done":    true,
	})

	BroadcastMessage("order", map[string]interface{}{"items": order, "churchName": cn})
	go saveDisplayState()

	// OBS: 첫 항목 씬 전환
	if len(order) > 0 {
		if title, ok := order[0]["title"].(string); ok {
			go obs.Get().SwitchScene(title)
		}
	}

	// 생성 이력 기록 (order 포함) — email 없으면 RecordGeneration 내부에서 "local@localhost" 사용
	go RecordGeneration(displayEmail, "display", time.Now().Format("060102"), "", "success", order)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "count": len(order)})
}

// DisplayNavigateHandler — POST /display/navigate
// 운영자 UI에서 원격 슬라이드 이동 (서브페이지 포함)
// 서버는 currentIdx를 업데이트하지 않음 — display HTML이 position 메시지로 보고
func DisplayNavigateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	var payload struct {
		Direction   string `json:"direction"`   // "next" | "prev" | "jump_sub"
		SubPageIdx  int    `json:"subPageIdx"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	msg := map[string]interface{}{"direction": payload.Direction}
	if payload.Direction == "jump_sub" {
		msg["subPageIdx"] = payload.SubPageIdx
	}
	log.Printf("[navigate] direction=%s broadcast to clients", payload.Direction)
	BroadcastMessage("navigate", msg)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

// DisplayPushHandler — POST /display/push (단독 항목 push, 기존 호환)
func DisplayPushHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	info, _ := payload["info"].(string)
	contents, _ := payload["contents"].(string)
	obj, _ := payload["obj"].(string)

	if strings.HasPrefix(info, "b_") && strings.TrimSpace(contents) == "" && obj != "" {
		versionID := 1
		if vid, ok := payload["versionId"].(float64); ok && vid > 0 {
			versionID = int(vid)
		}
		text, humanRef := fetchBibleTextWithVersion(obj, versionID)
		if text != "" {
			payload["contents"] = text
		}
		if humanRef != "" {
			payload["obj"] = humanRef
		}
	}

	BroadcastMessage("display", payload)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

// DisplayJumpHandler — POST /display/jump
// 순서 목록에서 특정 항목으로 점프
func DisplayJumpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	var payload struct {
		Index      int `json:"index"`
		SubPageIdx int `json:"subPageIdx"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	orderMu.Lock()
	if payload.Index < 0 || payload.Index >= len(currentOrder) {
		orderMu.Unlock()
		http.Error(w, "Index out of range", http.StatusBadRequest)
		return
	}
	currentIdx = payload.Index
	item := currentOrder[currentIdx]
	title, _ := item["title"].(string)
	info, _ := item["info"].(string)
	orderMu.Unlock()

	navPayload := map[string]interface{}{
		"direction":  "jump",
		"idx":        payload.Index,
		"subPageIdx": payload.SubPageIdx,
		"title":      title,
		"info":       info,
	}

	// OBS 씬 전환은 WS position 핸들러에서 통합 처리 (중복 방지)

	log.Printf("[jump] idx=%d title=%s broadcast to clients", payload.Index, title)
	BroadcastMessage("navigate", navPayload)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "idx": payload.Index})
}

// DisplayLyricsOrderHandler — POST /display/lyrics-order
// 가사 텍스트 → Display용 슬라이드 목록으로 변환하여 전송
func DisplayLyricsOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Songs []struct {
			Title  string `json:"title"`
			Lyrics string `json:"lyrics"`
			BPM    int    `json:"bpm"`
		} `json:"songs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	BroadcastMessage("display_loading", map[string]interface{}{
		"message": "찬양 가사 준비 중...",
		"total":   len(payload.Songs),
	})

	var order []map[string]interface{}
	for _, song := range payload.Songs {
		songMap := map[string]interface{}{
			"title":  song.Title,
			"lyrics": song.Lyrics,
			"bpm":    song.BPM,
		}
		order = append(order, preprocessLyricsItem(songMap))
	}

	// 새 순서 로드 시 타이머 초기화
	stopServerTimer()

	orderMu.Lock()
	currentOrder = order
	currentIdx = 0
	cn := displayChurchName
	orderMu.Unlock()

	BroadcastMessage("display_loading", map[string]interface{}{
		"message": "준비 완료!",
		"done":    true,
	})

	BroadcastMessage("order", map[string]interface{}{"items": order, "churchName": cn})
	go saveDisplayState()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "count": len(order)})
}

// lyricsSection — 가사 섹션 구조
type lyricsSection struct {
	Label string
	Text  string
}

// splitLyricsSections — 빈 줄 기준으로 가사를 섹션 분리 + 라벨 자동 부여
func splitLyricsSections(lyrics string) []lyricsSection {
	lines := strings.Split(strings.TrimSpace(lyrics), "\n")
	var sections []lyricsSection
	var current []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			if len(current) > 0 {
				sections = append(sections, lyricsSection{Text: strings.Join(current, "\n")})
				current = nil
			}
			continue
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		sections = append(sections, lyricsSection{Text: strings.Join(current, "\n")})
	}

	if len(sections) == 0 {
		return []lyricsSection{{Label: "전체", Text: lyrics}}
	}

	// 라벨 자동 부여: 반복 텍스트 감지
	textMap := make(map[string]int)
	for i := range sections {
		textMap[sections[i].Text]++
	}

	verseNum := 1
	chorusText := ""
	for i := range sections {
		if textMap[sections[i].Text] > 1 {
			// 반복되는 블록 → 후렴
			if chorusText == "" {
				chorusText = sections[i].Text
			}
			sections[i].Label = "후렴"
		} else {
			sections[i].Label = fmt.Sprintf("%d절", verseNum)
			verseNum++
		}
	}

	return sections
}

// DisplayAppendHandler — POST /display/append
// 기존 순서에 항목을 추가 (교체가 아닌 append)
func DisplayAppendHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Items    []map[string]interface{} `json:"items"`
		Source   string                   `json:"source"`
		AfterIdx *int                     `json:"afterIdx"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if len(payload.Items) == 0 {
		http.Error(w, "No items", http.StatusBadRequest)
		return
	}

	BroadcastMessage("display_loading", map[string]interface{}{
		"message": "항목 추가 준비 중...",
		"total":   len(payload.Items),
	})

	// 항목별 전처리 + source 태깅
	var processed []map[string]interface{}
	for i, item := range payload.Items {
		info, _ := item["info"].(string)
		title, _ := item["title"].(string)

		BroadcastMessage("display_loading", map[string]interface{}{
			"message": fmt.Sprintf("%s 처리 중... (%d/%d)", title, i+1, len(payload.Items)),
			"current": i + 1,
			"total":   len(payload.Items),
		})

		var result map[string]interface{}
		if info == "lyrics_display" {
			result = preprocessLyricsItem(item)
		} else {
			result = preprocessItem(item)
		}
		// source 태깅 (lyrics/bible)
		if payload.Source != "" {
			result["source"] = payload.Source
		}
		processed = append(processed, result)
	}

	orderMu.Lock()
	// afterIdx가 지정되면 해당 위치 뒤에 삽입, 아니면 끝에 추가
	if payload.AfterIdx != nil && *payload.AfterIdx >= 0 && *payload.AfterIdx < len(currentOrder) {
		insertAt := *payload.AfterIdx + 1
		tail := make([]map[string]interface{}, len(currentOrder[insertAt:]))
		copy(tail, currentOrder[insertAt:])
		currentOrder = append(currentOrder[:insertAt], processed...)
		currentOrder = append(currentOrder, tail...)
	} else {
		currentOrder = append(currentOrder, processed...)
	}
	order, idx, cn := getOrderSnapshotLocked()
	orderMu.Unlock()

	BroadcastMessage("display_loading", map[string]interface{}{
		"message": "준비 완료!",
		"done":    true,
	})

	BroadcastMessage("order", map[string]interface{}{"items": order, "idx": idx, "churchName": cn})
	go saveDisplayState()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "count": len(order)})
}

// DisplayRemoveHandler — POST /display/remove
// 순서 목록에서 특정 인덱스의 항목 제거
func DisplayRemoveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Index int `json:"index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	orderMu.Lock()
	if payload.Index < 0 || payload.Index >= len(currentOrder) {
		orderMu.Unlock()
		http.Error(w, "Index out of range", http.StatusBadRequest)
		return
	}
	currentOrder = append(currentOrder[:payload.Index], currentOrder[payload.Index+1:]...)
	// currentIdx 보정
	if payload.Index < currentIdx {
		currentIdx--
	} else if payload.Index == currentIdx && currentIdx >= len(currentOrder) && len(currentOrder) > 0 {
		currentIdx = len(currentOrder) - 1
	}
	if currentIdx < 0 {
		currentIdx = 0
	}
	order, idx, cn := getOrderSnapshotLocked()
	orderMu.Unlock()

	BroadcastMessage("order", map[string]interface{}{"items": order, "idx": idx, "churchName": cn})
	go saveDisplayState()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "count": len(order)})
}

// DisplayTimerHandler — POST /display/timer
// 제어판에서 서버 타이머 직접 제어
func DisplayTimerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	var payload struct {
		Action string  `json:"action"` // enable/disable/toggle/repeat/restart/speed
		Factor float64 `json:"factor"` // speed 조절 배율
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("[timer] action=%s factor=%.1f", payload.Action, payload.Factor)

	switch payload.Action {
	case "enable":
		timerMu.Lock()
		timerEnabled = true
		timerMu.Unlock()
		restartServerTimer()
	case "disable":
		stopServerTimer()
	case "toggle":
		timerMu.Lock()
		timerEnabled = !timerEnabled
		wasEnabled := timerEnabled
		timerMu.Unlock()
		if wasEnabled {
			restartServerTimer()
		} else {
			stopServerTimer()
		}
	case "repeat":
		// 현재 항목의 서브페이지 0으로 돌아가기
		timerMu.Lock()
		timerCurSubPage = -1 // force change detection
		timerMu.Unlock()
		BroadcastMessage("navigate", map[string]interface{}{
			"direction":  "jump_sub",
			"subPageIdx": 0,
		})
	case "restart":
		// 현재 항목의 처음으로
		orderMu.RLock()
		idx := currentIdx
		orderMu.RUnlock()
		timerMu.Lock()
		timerCurIdx = -1
		timerCurSubPage = -1
		timerMu.Unlock()
		BroadcastMessage("navigate", map[string]interface{}{
			"direction": "jump",
			"idx":       idx,
		})
	case "speed":
		if payload.Factor > 0 {
			timerMu.Lock()
			timerSpeedFactor = payload.Factor
			timerMu.Unlock()
			broadcastTimerState()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

// DisplayReorderHandler — POST /display/reorder
// 순서 목록에서 항목 위치 이동 (드래그 앤 드롭)
func DisplayReorderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		From int `json:"from"`
		To   int `json:"to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	orderMu.Lock()
	n := len(currentOrder)
	if payload.From < 0 || payload.From >= n || payload.To < 0 || payload.To >= n {
		orderMu.Unlock()
		http.Error(w, "Index out of range", http.StatusBadRequest)
		return
	}

	// 새 슬라이스에 복사 (원본 배열 변경 방지)
	newOrder := make([]map[string]interface{}, 0, n)
	item := currentOrder[payload.From]
	for i, v := range currentOrder {
		if i == payload.From {
			continue
		}
		if len(newOrder) == payload.To {
			newOrder = append(newOrder, item)
		}
		newOrder = append(newOrder, v)
	}
	if len(newOrder) == payload.To {
		newOrder = append(newOrder, item)
	}
	currentOrder = newOrder

	// currentIdx 보정
	if currentIdx == payload.From {
		currentIdx = payload.To
	} else {
		if payload.From < currentIdx {
			currentIdx--
		}
		if payload.To <= currentIdx {
			currentIdx++
		}
	}
	if currentIdx < 0 {
		currentIdx = 0
	} else if currentIdx >= n {
		currentIdx = n - 1
	}

	order, idx, cn := getOrderSnapshotLocked()
	orderMu.Unlock()

	BroadcastMessage("order", map[string]interface{}{"items": order, "idx": idx, "churchName": cn})
	go saveDisplayState()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "count": len(order)})
}

// DisplayChurchNameHandler — POST /display/church-name
// 교회명 변경 시 Display에 즉시 반영
func DisplayChurchNameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		ChurchName string `json:"churchName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	orderMu.Lock()
	displayChurchName = body.ChurchName
	order, idx, cn := getOrderSnapshotLocked()
	orderMu.Unlock()

	BroadcastMessage("order", map[string]interface{}{"items": order, "idx": idx, "churchName": cn})
	go saveDisplayState()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

// DisplayStatusHandler — GET /display/status
// 현재 상태 반환 (idx, 항목 목록, OBS 상태)
func DisplayStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	orderMu.RLock()
	items, idx, _ := getOrderSnapshotLocked()
	count := len(items)

	var title string
	if idx >= 0 && idx < count {
		title, _ = items[idx]["title"].(string)
	}
	orderMu.RUnlock()

	obsStatus := obs.Get().GetStatus()

	streamStatus := obs.Get().GetStreamStatus()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"idx":    idx,
		"count":  count,
		"title":  title,
		"items":  items,
		"obs":    obsStatus,
		"stream": streamStatus,
	})
}

// preprocessItem — 단일 항목 전처리 (성경, 신앙고백, 주기도문, 교회소식, 찬송/교독 이미지)
func preprocessItem(item map[string]interface{}) map[string]interface{} {
	info, _ := item["info"].(string)
	obj, _ := item["obj"].(string)
	title, _ := item["title"].(string)

	// b_edit: 성경 본문 자동 조회
	if strings.HasPrefix(info, "b_") && obj != "" {
		versionID := 1
		if vid, ok := item["versionId"].(float64); ok && vid > 0 {
			versionID = int(vid)
		}
		text, humanRef := fetchBibleTextWithVersion(obj, versionID)
		if text != "" {
			item["contents"] = text
		}
		if humanRef != "" {
			item["obj"] = humanRef
		}
	}

	// 신앙고백 (사도신경): 본문 자동 삽입
	if title == "신앙고백" {
		item["contents"] = apostlesCreed
	}

	// 주기도문: 본문 자동 삽입
	if title == "주기도문" {
		item["contents"] = lordsPrayer
	}

	// 항목별 배경 이미지 — data/templates/display/{title}.png/.jpg 자동 매핑
	{
		execPath := path.ExecutePath("easyPreparation")
		displayDir := filepath.Join(execPath, "data", "templates", "display")
		for _, ext := range []string{".png", ".jpg", ".jpeg"} {
			bgPath := filepath.Join(displayDir, title+ext)
			if _, err := os.Stat(bgPath); err == nil {
				item["bgImage"] = "/display/assets/" + url.PathEscape(title+ext)
				break
			}
		}
	}

	// 교회소식: children → contents 계층 텍스트 전처리
	if info == "notice" || strings.Contains(title, "교회소식") {
		if rawChildren, ok := item["children"]; ok {
			childrenJSON, _ := json.Marshal(rawChildren)
			var children []map[string]interface{}
			if json.Unmarshal(childrenJSON, &children) == nil && len(children) > 0 {
				item["contents"] = formatChurchNews(children, 1)
			}
		}
	}

	// 찬송/헌금봉헌/성시교독: Google Drive PDF → PNG 변환
	if title == "찬송" || title == "헌금봉헌" || title == "성시교독" {
		if images := fetchDisplayImages(title, obj); len(images) > 0 {
			item["images"] = images

			// 찬송: 가사↔페이지 자동 매핑
			if title == "찬송" || title == "헌금봉헌" {
				var hymnNum int
				for _, r := range obj {
					if r >= '0' && r <= '9' {
						hymnNum = hymnNum*10 + int(r-'0')
					}
				}
				if hymnNum > 0 {
					if lyrics := fetchHymnLyrics(hymnNum); lyrics != "" {
						verses := splitVerses(lyrics)
						if lm := mapLyricsToPages(verses, len(images)); len(lm) > 0 {
							item["lyricsMap"] = lm
						}
					}
				}
			}
		}
	}

	// ── sections 생성 (제어판 토글용) ──
	// Display HTML의 서브페이지 분할 로직과 동일하게 맞춤
	item = buildSections(item)

	return item
}

// buildSections — 항목의 서브페이지 구조에 맞춰 sections 배열 생성
func buildSections(item map[string]interface{}) map[string]interface{} {
	info, _ := item["info"].(string)
	title, _ := item["title"].(string)
	contents, _ := item["contents"].(string)

	// 이미 sections가 있으면 스킵 (lyrics 등)
	if _, ok := item["sections"]; ok {
		return item
	}

	// 1. 성경 본문 → 3줄 단위 페이징
	if strings.HasPrefix(info, "b_") && contents != "" {
		pages := paginateText(contents, 3)
		if len(pages) > 1 {
			item["sections"] = buildTextSections(pages)
		}
		return item
	}

	// 2. 신앙고백 → 10줄 단위 페이징 (주기도문은 짧아서 분할 불필요)
	if title == "신앙고백" && contents != "" {
		pages := paginateText(contents, 10)
		if len(pages) > 1 {
			item["sections"] = buildTextSections(pages)
		}
		return item
	}

	// 3. 찬송/헌금봉헌/성시교독 → 이미지 페이지 (성시교독은 표지 없음)
	if title == "찬송" || title == "헌금봉헌" || title == "성시교독" {
		var images []string
		switch v := item["images"].(type) {
		case []string:
			images = v
		case []interface{}:
			for _, img := range v {
				if s, ok := img.(string); ok {
					images = append(images, s)
				}
			}
		}
		if len(images) > 0 {
			obj, _ := item["obj"].(string)
			// lyricsMap이 있으면 각 이미지 페이지에 가사 미리보기 삽입
			var lyricsMap []string
			if lm, ok := item["lyricsMap"].([]string); ok {
				lyricsMap = lm
			}
			var sections []map[string]interface{}
			if title == "성시교독" {
				// 성시교독: 표지 없이 이미지만
				for i := range images {
					sections = append(sections, map[string]interface{}{
						"label":     fmt.Sprintf("%d", i+1),
						"startPage": i,
						"text":      "",
					})
				}
			} else {
				// 찬송/헌금봉헌: 표지 + 이미지
				sections = []map[string]interface{}{
					{"label": "표지", "startPage": 0, "text": obj},
				}
				for i := range images {
					preview := ""
					if i < len(lyricsMap) && lyricsMap[i] != "" {
						preview = lyricsMap[i]
						runes := []rune(preview)
						if len(runes) > 60 {
							preview = string(runes[:60]) + "..."
						}
						preview = strings.ReplaceAll(preview, "\n", " ")
					}
					sections = append(sections, map[string]interface{}{
						"label":     fmt.Sprintf("%d", i+1),
						"startPage": i + 1,
						"text":      preview,
					})
				}
			}
			item["sections"] = sections
		}
		return item
	}

	return item
}

// paginateText — 빈 줄 제거 후 N줄 단위로 페이지 분할 (Display HTML의 paginate 함수와 동일)
func paginateText(text string, linesPerPage int) []string {
	allLines := strings.Split(text, "\n")
	var lines []string
	for _, l := range allLines {
		if strings.TrimSpace(l) != "" {
			lines = append(lines, l)
		}
	}
	if len(lines) == 0 {
		return []string{text}
	}
	var pages []string
	for i := 0; i < len(lines); i += linesPerPage {
		end := i + linesPerPage
		if end > len(lines) {
			end = len(lines)
		}
		pages = append(pages, strings.Join(lines[i:end], "\n"))
	}
	return pages
}

// buildTextSections — 텍스트 페이지 배열 → sections 배열
func buildTextSections(pages []string) []map[string]interface{} {
	sections := make([]map[string]interface{}, len(pages))
	for i, page := range pages {
		// 미리보기: 첫 줄만 표시
		preview := page
		if idx := strings.Index(page, "\n"); idx > 0 {
			preview = page[:idx] + " ..."
		}
		sections[i] = map[string]interface{}{
			"label":     fmt.Sprintf("%d", i+1),
			"startPage": i,
			"text":      preview,
		}
	}
	return sections
}

// preprocessLyricsItem — 가사 곡 하나를 Display 항목으로 변환
func preprocessLyricsItem(song map[string]interface{}) map[string]interface{} {
	title, _ := song["title"].(string)
	lyrics, _ := song["lyrics"].(string)
	bpm := 0
	switch v := song["bpm"].(type) {
	case float64:
		bpm = int(v)
	case int:
		bpm = v
	}

	rawLines := strings.Split(strings.TrimSpace(lyrics), "\n")
	var lines []string
	for _, l := range rawLines {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			lines = append(lines, trimmed)
		}
	}

	// (xN) 표기는 독립 페이지가 아닌 이전 페이지에 붙임
	isRepeatMark := func(s string) bool {
		t := strings.TrimSpace(s)
		return len(t) >= 3 && t[0] == '(' && t[1] == 'x' && t[len(t)-1] == ')'
	}

	var allPages []string
	var sectionMarkers []map[string]interface{}
	for i := 0; i < len(lines); i += 2 {
		end := i + 2
		if end > len(lines) {
			end = len(lines)
		}
		pageText := strings.Join(lines[i:end], "\n")

		// 다음 줄이 (xN)이면 현재 페이지에 포함
		if end < len(lines) && isRepeatMark(lines[end]) {
			pageText += "\n" + lines[end]
			i++ // 다음 루프에서 (xN)줄 건너뜀
		}

		allPages = append(allPages, pageText)
		sectionMarkers = append(sectionMarkers, map[string]interface{}{
			"label":     fmt.Sprintf("%d", len(sectionMarkers)+1),
			"startPage": len(sectionMarkers),
			"text":      pageText,
		})
	}

	return map[string]interface{}{
		"title":    title,
		"info":     "lyrics_display",
		"obj":      "-",
		"contents": lyrics,
		"bpm":      bpm,
		"pages":    allPages,
		"sections": sectionMarkers,
	}
}

// formatChurchNews — 교회소식 children 재귀 순회하여 계층 텍스트 생성
func formatChurchNews(items []map[string]interface{}, depth int) string {
	var sb strings.Builder
	romanNumerals := []string{"i", "ii", "iii", "iv", "v", "vi", "vii", "viii", "ix", "x"}
	for i, item := range items {
		title, _ := item["title"].(string)
		obj, _ := item["obj"].(string)
		if obj == "-" {
			obj = ""
		}
		tab := strings.Repeat("    ", depth-1)
		// depth별 인덱스 포맷
		var index string
		switch depth {
		case 1:
			index = fmt.Sprintf("%d.", i+1)
		case 2:
			index = fmt.Sprintf("%d)", i+1)
		case 3:
			if i < 26 {
				index = fmt.Sprintf("%c)", 'a'+i)
			}
		default:
			if i < len(romanNumerals) {
				index = fmt.Sprintf("%s)", romanNumerals[i])
			}
		}
		// children 유무에 따라 title 포맷
		hasChildren := false
		if rawChildren, ok := item["children"]; ok {
			childJSON, _ := json.Marshal(rawChildren)
			var ch []map[string]interface{}
			if json.Unmarshal(childJSON, &ch) == nil && len(ch) > 0 {
				hasChildren = true
			}
		}
		displayTitle := title
		if !hasChildren && title != "" {
			displayTitle = title + ":"
		}
		line := strings.TrimRight(fmt.Sprintf("%s%s %s %s", tab, index, displayTitle, obj), " ")
		sb.WriteString(line + "\n")
		// 재귀 처리
		if hasChildren {
			childJSON, _ := json.Marshal(item["children"])
			var ch []map[string]interface{}
			_ = json.Unmarshal(childJSON, &ch)
			sb.WriteString(formatChurchNews(ch, depth+1))
		}
	}
	return sb.String()
}

// fetchDisplayImages — 찬송/교독문 PDF를 Google Drive에서 다운로드 → PNG 변환 → URL 목록 반환
func fetchDisplayImages(title, obj string) []string {
	var category, splitNum string
	switch title {
	case "찬송", "헌금봉헌":
		category = "hymn"
		for _, r := range obj {
			if r >= '0' && r <= '9' {
				splitNum += string(r)
			}
		}
		if splitNum == "" {
			splitNum = obj
		}
	case "성시교독":
		category = "responsive_reading"
		splitNum = strings.Split(obj, ".")[0]
	default:
		return nil
	}

	execPath := path.ExecutePath("easyPreparation")
	baseDir := filepath.Join(execPath, "data", "cache", "hymn_pages")
	_ = utils.CheckDirIs(baseDir)

	// 숫자 0-패딩
	num, err := strconv.Atoi(splitNum)
	var targetNum string
	if err == nil {
		targetNum = fmt.Sprintf("%03d.pdf", num)
	} else {
		targetNum = fmt.Sprintf("%s.pdf", splitNum)
	}

	// 캐시: 이미 변환된 PNG가 있으면 재사용
	cachePrefix := fmt.Sprintf("%s_%s_", category, strings.TrimSuffix(targetNum, ".pdf"))
	if cached := findCachedImages(baseDir, cachePrefix); len(cached) > 0 {
		return cached
	}

	// 고정 디렉토리에 PDF 캐시 (data/pdf/hymn/, data/pdf/responsive_reading/)
	pdfDir := filepath.Join(execPath, "data", "pdf")
	_ = utils.CheckDirIs(pdfDir)
	cacheDir := filepath.Join(pdfDir, category)
	_ = utils.CheckDirIs(cacheDir)

	pdfPath := filepath.Join(cacheDir, targetNum)

	// PDF 캐시 확인 → 없으면 R2에서 다운로드
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		BroadcastMessage("display_loading", map[string]interface{}{
			"message": fmt.Sprintf("%s/%s 다운로드 중...", category, targetNum),
		})
		if err := assets.DownloadPDF(category, targetNum, cacheDir); err != nil {
			log.Printf("[display] PDF 다운로드 실패 — %v (건너뜀)", err)
			return nil
		}
	} else {
		BroadcastMessage("display_loading", map[string]interface{}{
			"message": fmt.Sprintf("[캐시] %s/%s 사용", category, targetNum),
		})
	}

	// PDF → PNG 변환 (gs)
	tmpDir, err := os.MkdirTemp(baseDir, category+"_conv_")
	if err != nil {
		log.Printf("[display] 임시 디렉토리 생성 실패: %v", err)
		return nil
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()
	pngPattern := filepath.Join(tmpDir, "page_%d.png")

	gsPath := "/opt/homebrew/bin/gs"
	if _, err := os.Stat(gsPath); err != nil {
		gsPath = "gs" // fallback to PATH
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("gswin64c", "-sDEVICE=pngalpha", "-o", pngPattern, "-r144", pdfPath)
	} else {
		cmd = exec.Command(gsPath, "-sDEVICE=pngalpha", "-o", pngPattern, "-r144", pdfPath)
	}
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("[display] 찬송/교독 변환 실패: %s, 에러: %v", string(output), err)
		return nil
	}

	// PNG 파일 → baseDir에 고유명으로 복사
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		return nil
	}
	var pngFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".png") {
			pngFiles = append(pngFiles, f.Name())
		}
	}
	sort.Strings(pngFiles)

	var urls []string
	for _, pf := range pngFiles {
		dstName := cachePrefix + pf
		src := filepath.Join(tmpDir, pf)
		dst := filepath.Join(baseDir, dstName)
		if data, err := os.ReadFile(src); err == nil {
			_ = os.WriteFile(dst, data, 0644)
			urls = append(urls, "/display/tmp/"+dstName)
		}
	}
	return urls
}

// findCachedImages — baseDir에서 prefix로 시작하는 PNG 파일 URL 목록 반환
func findCachedImages(baseDir, prefix string) []string {
	files, err := os.ReadDir(baseDir)
	if err != nil {
		return nil
	}
	var urls []string
	for _, f := range files {
		if !f.IsDir() && strings.HasPrefix(f.Name(), prefix) && strings.HasSuffix(f.Name(), ".png") {
			urls = append(urls, "/display/tmp/"+f.Name())
		}
	}
	sort.Strings(urls)
	return urls
}

// ── 가사↔페이지 매핑 ──

// fetchHymnLyrics — hymns 테이블에서 가사 텍스트 조회 (PostgreSQL bibleDB)
func fetchHymnLyrics(number int) string {
	if bibleDB == nil {
		return ""
	}
	var lyrics string
	err := bibleDB.QueryRow("SELECT lyrics FROM hymns WHERE hymnbook='new' AND number=?", number).Scan(&lyrics)
	if err != nil {
		return ""
	}
	return lyrics
}

var versePattern = regexp.MustCompile(`(?m)^\d+\.\s`)

// splitVerses — 가사를 절 단위로 분리
func splitVerses(lyrics string) []string {
	lyrics = strings.TrimSpace(lyrics)
	if lyrics == "" {
		return nil
	}

	// 절 번호 패턴 (1. 2. ...) 으로 분리 시도
	locs := versePattern.FindAllStringIndex(lyrics, -1)
	if len(locs) >= 2 {
		var verses []string
		for i, loc := range locs {
			start := loc[0]
			var end int
			if i+1 < len(locs) {
				end = locs[i+1][0]
			} else {
				end = len(lyrics)
			}
			verse := strings.TrimSpace(lyrics[start:end])
			if verse != "" {
				verses = append(verses, verse)
			}
		}
		return verses
	}

	// 절 번호 없으면 빈 줄(\n\n)로 분리
	blocks := strings.Split(lyrics, "\n\n")
	var verses []string
	for _, b := range blocks {
		b = strings.TrimSpace(b)
		if b != "" {
			verses = append(verses, b)
		}
	}
	if len(verses) <= 1 {
		return []string{lyrics}
	}
	return verses
}

// splitIntoChunks — 텍스트를 N줄 단위로 분할
func splitIntoChunks(text string, linesPerChunk int) []string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	var chunks []string
	for i := 0; i < len(lines); i += linesPerChunk {
		end := i + linesPerChunk
		if end > len(lines) {
			end = len(lines)
		}
		chunks = append(chunks, strings.Join(lines[i:end], "\n"))
	}
	if len(chunks) == 0 {
		return []string{text}
	}
	return chunks
}

// mapLyricsToPages — 전체 가사를 페이지 수에 맞게 빠짐없이 균등 분배
func mapLyricsToPages(verses []string, pageCount int) []string {
	if pageCount <= 0 || len(verses) == 0 {
		return nil
	}

	// 전체 가사를 줄 단위로 펼침
	var allLines []string
	for _, v := range verses {
		for _, line := range strings.Split(strings.TrimSpace(v), "\n") {
			if strings.TrimSpace(line) != "" {
				allLines = append(allLines, line)
			}
		}
	}

	n := len(allLines)
	if n == 0 {
		return nil
	}

	// 각 페이지에 연속된 줄을 균등 배분
	result := make([]string, pageCount)
	for i := 0; i < pageCount; i++ {
		start := i * n / pageCount
		end := (i + 1) * n / pageCount
		if start >= n {
			start = n - 1
		}
		if end > n {
			end = n
		}
		result[i] = strings.Join(allLines[start:end], "\n")
	}
	return result
}

// fetchBibleText — "책명_코드/장:절" 형식에서 성경 본문 조회 (기본 버전)
func fetchBibleText(obj string) (string, string) {
	return fetchBibleTextWithVersion(obj, 1)
}

// fetchBibleTextWithVersion — 지정된 버전으로 성경 본문 조회
func fetchBibleTextWithVersion(obj string, versionID int) (string, string) {
	var texts []string
	var refs []string

	for _, part := range strings.Split(obj, ",") {
		part = strings.TrimSpace(part)
		underIdx := strings.Index(part, "_")
		if underIdx < 0 {
			continue
		}
		korName := part[:underIdx]
		codeAndRange := part[underIdx+1:]

		slashIdx := strings.Index(codeAndRange, "/")
		if slashIdx < 0 {
			continue
		}
		verseRange := codeAndRange[slashIdx+1:]

		text, err := quote.GetQuoteWithVersion(codeAndRange, versionID)
		if err != nil {
			continue
		}
		texts = append(texts, text)
		refs = append(refs, fmt.Sprintf("%s %s", korName, verseRange))
	}

	return strings.Join(texts, "\n"), strings.Join(refs, ", ")
}
