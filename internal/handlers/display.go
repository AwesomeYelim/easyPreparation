package handlers

import (
	"easyPreparation_1.0/internal/googleCloud"
	"easyPreparation_1.0/internal/obs"
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/internal/quote"
	"easyPreparation_1.0/internal/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// 현재 예배 순서 메모리 저장
var (
	orderMu      sync.RWMutex
	currentOrder []map[string]interface{}
	currentIdx   int
)

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

</style>
</head>
<body>
<div id="slide"></div>

<script>
const slide = document.getElementById('slide');
let ws, reconnectTimer;
let slides = [];
let idx = 0;
let subPages = [];   // 성경 본문 or 이미지 페이지
let subPageIdx = 0;

/* ───── WebSocket ───── */
function connect() {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws';
  ws = new WebSocket(proto + '://' + location.host + '/ws');
  ws.onmessage = (e) => {
    const msg = JSON.parse(e.data);
    if (msg.type === 'order') loadOrder(msg.items, msg.idx);
    if (msg.type === 'navigate') {
      if (msg.direction === 'jump' && typeof msg.idx === 'number') {
        showSlide(msg.idx);
      } else {
        navigate(msg.direction);
      }
    }
    if (msg.type === 'display') renderSingle(msg);
  };
  ws.onclose = () => {
    clearTimeout(reconnectTimer);
    reconnectTimer = setTimeout(connect, 3000);
  };
}

/* ───── 순서 로드 ───── */
function loadOrder(items, startIdx) {
  slides = items || [];
  idx = 0;
  var start = (typeof startIdx === 'number') ? startIdx : 0;
  showSlide(start);
}

/* ───── 슬라이드 표시 ───── */
let lastReportedIdx = -1;
function showSlide(i) {
  if (!slides.length) return;
  i = Math.max(0, Math.min(i, slides.length - 1));
  idx = i;
  // 항목 변경 시 서버에 위치 보고
  if (idx !== lastReportedIdx) {
    lastReportedIdx = idx;
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({type:'position', idx:idx}));
    }
  }
  const item = slides[idx];

  subPages = [];
  subPageIdx = 0;

  const itemTitle = item.title || '';

  // 성경 본문 → 텍스트 페이지 분할
  if ((item.info || '').startsWith('b_') && item.contents) {
    subPages = paginate(item.contents, 5);
  }
  // 신앙고백 본문 → 페이지 분할
  else if (itemTitle === '신앙고백' && item.contents) {
    subPages = paginate(item.contents, 10);
  }
  // 찬송/교독 이미지 → 표지 + 이미지 페이지
  else if (item.images && item.images.length > 0) {
    subPages = ['__cover__'].concat(item.images);
  }

  renderItem(item, 0);
}

/* ───── 키보드 / 네비게이션 ───── */
function navigate(dir) {
  if (dir === 'next') {
    if (subPages.length > 1 && subPageIdx < subPages.length - 1) {
      subPageIdx++;
      renderItem(slides[idx], subPageIdx);
    } else {
      showSlide(idx + 1);
    }
  } else if (dir === 'prev') {
    if (subPages.length > 1 && subPageIdx > 0) {
      subPageIdx--;
      renderItem(slides[idx], subPageIdx);
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
  const info     = item.info     || '';
  const title    = item.title    || '';
  const obj      = item.obj      || '';
  const lead     = item.lead     || '';
  const contents = item.contents || '';
  const images   = item.images   || [];

  slide.style.backgroundImage = "linear-gradient(rgba(0,0,0,0.4),rgba(0,0,0,0.4)), url('/display/bg')";
  slide.className = 'visible';

  const posText = slides.length ? (idx + 1) + ' / ' + slides.length : '';
  const pageText = subPages.length > 1 ? (subPageIdx + 1) + ' / ' + subPages.length : '';
  const footer =
    '<div class="divider"></div>' +
    '<div class="slide-pos">' + posText + '</div>' +
    (pageText ? '<div class="page-indicator">' + pageText + '</div>' : '');
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

  // ── 3. 성시교독 (표지 + 이미지 페이지) ──
  if (title === '성시교독') {
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

  // ── 6. 참회의 기도 (좌정렬 멀티라인) ──
  if (title === '참회의 기도') {
    slide.innerHTML = header +
      '<div class="title">' + esc(title) + '</div>' +
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

  // ── 8. 기본 (전주, 개회기도, 신앙고백, 기도, 축도 등) ──
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

// DisplayHandler — GET /display
func DisplayHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	_, _ = w.Write([]byte(displayHTML))
}

// DisplayAssetsHandler — GET /display/assets/{name}.png
// Figma 배경 이미지 서빙
func DisplayAssetsHandler(w http.ResponseWriter, r *http.Request) {
	name := filepath.Base(r.URL.Path)
	execPath := path.ExecutePath("easyPreparation")
	imgPath := filepath.Join(execPath, "output", "bulletin", "presentation", "tmp", name)
	http.ServeFile(w, r, imgPath)
}

// DisplayBgHandler — GET /display/bg
// 공통 배경 이미지 서빙
func DisplayBgHandler(w http.ResponseWriter, r *http.Request) {
	execPath := path.ExecutePath("easyPreparation")
	imgPath := filepath.Join(execPath, "output", "lyrics", "tmp", "Frame 1.png")
	http.ServeFile(w, r, imgPath)
}

// DisplayTmpHandler — GET /display/tmp/{name}
// 찬송/교독 변환 PNG 서빙
func DisplayTmpHandler(w http.ResponseWriter, r *http.Request) {
	name := filepath.Base(r.URL.Path)
	execPath := path.ExecutePath("easyPreparation")
	imgPath := filepath.Join(execPath, "output", "display", "tmp", name)
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
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	BroadcastMessage("display_loading", map[string]interface{}{
		"message": "예배 화면 준비 중...",
		"total":   len(order),
	})

	// 항목별 전처리
	for i, item := range order {
		info, _ := item["info"].(string)
		obj, _ := item["obj"].(string)
		title, _ := item["title"].(string)

		BroadcastMessage("display_loading", map[string]interface{}{
			"message": fmt.Sprintf("%s 처리 중... (%d/%d)", title, i+1, len(order)),
			"current": i + 1,
			"total":   len(order),
		})

		// b_edit: 성경 본문 자동 조회
		if strings.HasPrefix(info, "b_") && obj != "" {
			text, humanRef := fetchBibleText(obj)
			if text != "" {
				order[i]["contents"] = text
			}
			if humanRef != "" {
				order[i]["obj"] = humanRef
			}
		}

		// 신앙고백 (사도신경): 본문 자동 삽입
		if title == "신앙고백" {
			order[i]["contents"] = apostlesCreed
		}

		// 교회소식: children → contents 계층 텍스트 전처리
		if info == "notice" || strings.Contains(title, "교회소식") {
			if rawChildren, ok := item["children"]; ok {
				childrenJSON, _ := json.Marshal(rawChildren)
				var children []map[string]interface{}
				if json.Unmarshal(childrenJSON, &children) == nil && len(children) > 0 {
					order[i]["contents"] = formatChurchNews(children, 1)
				}
			}
		}

		// 찬송/헌금봉헌/성시교독: Google Drive PDF → PNG 변환
		if title == "찬송" || title == "헌금봉헌" || title == "성시교독" {
			if images := fetchDisplayImages(title, obj); len(images) > 0 {
				order[i]["images"] = images
			}
		}
	}

	orderMu.Lock()
	currentOrder = order
	currentIdx = 0
	orderMu.Unlock()

	BroadcastMessage("display_loading", map[string]interface{}{
		"message": "준비 완료!",
		"done":    true,
	})

	BroadcastMessage("order", map[string]interface{}{"items": order})

	// OBS: 첫 항목 씬 전환
	if len(order) > 0 {
		if title, ok := order[0]["title"].(string); ok {
			go obs.Get().SwitchScene(title)
		}
	}

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
		Direction string `json:"direction"` // "next" | "prev"
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	BroadcastMessage("navigate", map[string]interface{}{"direction": payload.Direction})
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
		text, humanRef := fetchBibleText(obj)
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
	currentIdx = payload.Index
	item := currentOrder[currentIdx]
	title, _ := item["title"].(string)
	info, _ := item["info"].(string)
	orderMu.Unlock()

	navPayload := map[string]interface{}{
		"direction": "jump",
		"idx":       payload.Index,
		"title":     title,
		"info":      info,
	}

	// OBS 씬 전환
	if title != "" {
		go obs.Get().SwitchScene(title)
	}

	BroadcastMessage("navigate", navPayload)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "idx": payload.Index})
}

// DisplayStatusHandler — GET /display/status
// 현재 상태 반환 (idx, 항목 목록, OBS 상태)
func DisplayStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	orderMu.RLock()
	idx := currentIdx
	count := len(currentOrder)

	var title string
	var items []map[string]interface{}
	for _, item := range currentOrder {
		t, _ := item["title"].(string)
		o, _ := item["obj"].(string)
		l, _ := item["lead"].(string)
		items = append(items, map[string]interface{}{
			"title": t, "obj": o, "lead": l,
		})
	}
	if idx >= 0 && idx < count {
		title, _ = currentOrder[idx]["title"].(string)
	}
	orderMu.RUnlock()

	obsStatus := obs.Get().GetStatus()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"idx":   idx,
		"count": count,
		"title": title,
		"items": items,
		"obs":   obsStatus,
	})
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
	baseDir := filepath.Join(execPath, "output", "display", "tmp")
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

	// 고정 디렉토리에 PDF 캐시 (data/hymn/, data/responsive_reading/)
	dataDir := filepath.Join(execPath, "data")
	_ = utils.CheckDirIs(dataDir)
	cacheDir := filepath.Join(dataDir, category)
	_ = utils.CheckDirIs(cacheDir)

	pdfPath := filepath.Join(cacheDir, targetNum)

	// PDF 캐시 확인 → 없으면 Google Drive에서 다운로드
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		BroadcastMessage("display_loading", map[string]interface{}{
			"message": fmt.Sprintf("Google Drive: %s/%s 다운로드 중...", category, targetNum),
		})
		if err := googleCloud.GetGoogleCloudInfo(category, targetNum, cacheDir); err != nil {
			log.Printf("[display] Google Drive 파일 없음 — %v (건너뜀)", err)
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

// fetchBibleText — "책명_코드/장:절" 형식에서 성경 본문 조회
func fetchBibleText(obj string) (string, string) {
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

		text, err := quote.GetQuote(codeAndRange)
		if err != nil {
			continue
		}
		texts = append(texts, text)
		refs = append(refs, fmt.Sprintf("%s %s", korName, verseRange))
	}

	return strings.Join(texts, "\n"), strings.Join(refs, ", ")
}
