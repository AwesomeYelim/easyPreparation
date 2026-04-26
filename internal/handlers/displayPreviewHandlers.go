package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// DisplayPreviewHandler — GET /display/preview?index=N
// 씬 패널 미리보기: WebSocket 없이 단일 항목을 Display와 동일한 스타일로 렌더링
func DisplayPreviewHandler(w http.ResponseWriter, r *http.Request) {
	indexStr := r.URL.Query().Get("index")
	index, _ := strconv.Atoi(indexStr)

	orderMu.RLock()
	order := deepCopyOrder(currentOrder)
	orderMu.RUnlock()

	if order == nil {
		order = []map[string]interface{}{}
	}
	if index < 0 || index >= len(order) {
		index = 0
	}

	allJSON, _ := json.Marshal(order)

	html := strings.Replace(displayPreviewHTML, "/*__SLIDES__*/[]", "/*__SLIDES__*/"+string(allJSON), 1)
	html = strings.Replace(html, "/*__IDX__*/0", "/*__IDX__*/"+strconv.Itoa(index), 1)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(html))
}

const displayPreviewHTML = `<!DOCTYPE html>
<html lang="ko">
<head>
<meta charset="UTF-8">
<title>Preview</title>
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
    z-index:1;
    background-size:cover;
    background-position:center;
    background-repeat:no-repeat;
    opacity:0;
    transition:opacity 0.15s ease;
  }
  #slide.visible { opacity:1; }
  .label { position:absolute; top:4.4vh; left:4.2vw; font-size:2.6vh; color:rgba(255,255,255,0.5); letter-spacing:0.05em; }
  .order-title { position:absolute; top:4.4vh; right:4.2vw; font-size:2.6vh; color:rgba(255,255,255,0.5); }
  .page-indicator { position:absolute; bottom:3vh; right:4.2vw; font-size:2vh; color:rgba(255,255,255,0.35); }
  .title { font-size:5.9vh; font-weight:700; text-align:center; margin-bottom:3.3vh; line-height:1.3; letter-spacing:-0.01em; text-shadow:0 2px 8px rgba(0,0,0,0.6); }
  .obj { font-size:4.4vh; font-weight:400; color:rgba(255,255,255,0.85); text-align:center; line-height:1.5; text-shadow:0 2px 6px rgba(0,0,0,0.5); }
  .bible-ref { font-size:3.5vh; color:rgba(255,255,255,0.65); margin-bottom:3vh; text-align:left; width:100%; text-shadow:0 1px 4px rgba(0,0,0,0.5); }
  .bible-contents { font-size:5vh; line-height:1.9; text-align:left; color:#fff; white-space:pre-wrap; width:100%; text-shadow:0 1px 6px rgba(0,0,0,0.6); }
  .hymn-number { font-size:10vh; font-weight:700; text-align:center; line-height:1.2; text-shadow:0 3px 12px rgba(0,0,0,0.7); }
  .hymn-sub { font-size:3.2vh; color:rgba(255,255,255,0.6); margin-top:2vh; text-align:center; }
  .slide-image { max-width:90vw; max-height:85vh; object-fit:contain; }
  .prayer-name { font-size:8vh; font-weight:700; text-align:center; line-height:1.3; text-shadow:0 3px 12px rgba(0,0,0,0.7); }
  .lyrics-text { font-size:7vh; line-height:1.8; text-align:center; color:#fff; white-space:pre-wrap; width:100%; font-weight:500; text-shadow:0 2px 10px rgba(0,0,0,0.7); }
  .creed-text { font-size:3.3vh; line-height:1.9; text-align:center; color:rgba(255,255,255,0.9); white-space:pre-wrap; width:100%; text-shadow:0 1px 6px rgba(0,0,0,0.5); }
  .confession-text { font-size:3.6vh; line-height:2; text-align:left; color:rgba(255,255,255,0.9); white-space:pre-wrap; width:100%; text-shadow:0 1px 6px rgba(0,0,0,0.5); }
  .sermon-title { font-size:7vh; font-weight:700; text-align:center; line-height:1.3; margin-bottom:3vh; text-shadow:0 3px 12px rgba(0,0,0,0.7); }
  .sermon-pastor { font-size:4vh; color:rgba(255,255,255,0.7); text-align:center; text-shadow:0 2px 6px rgba(0,0,0,0.5); }
  .notice-title { font-size:4.8vh; font-weight:700; margin-bottom:2.6vh; color:#fff; text-shadow:0 2px 8px rgba(0,0,0,0.6); }
  .notice-contents { font-size:2.8vh; color:rgba(255,255,255,0.85); text-align:left; line-height:1.8; white-space:pre-wrap; width:100%; text-shadow:0 1px 4px rgba(0,0,0,0.5); }
  .divider { position:absolute; bottom:5.6vh; left:4.2vw; right:4.2vw; height:1px; background:rgba(255,255,255,0.15); }
  .slide-pos { position:absolute; bottom:3vh; left:4.2vw; font-size:2vh; color:rgba(255,255,255,0.35); }
</style>
</head>
<body>
<video id="bg-video" autoplay loop muted playsinline
  style="position:fixed;top:0;left:0;width:100%;height:100%;object-fit:cover;z-index:0;display:none">
  <source id="bg-video-src" src="" type="video/mp4">
</video>
<div id="slide"></div>
<script>
const slide = document.getElementById('slide');
var slides = /*__SLIDES__*/[];
var idx    = /*__IDX__*/0;
var subPageIdx = 0;
var subPages   = [];
var logoUrl    = '';
var logoPosition     = 'bottom-right';
var logoSizePercent  = 18;
var activeVideoBg    = '';
var globalImageBgDisabled = false;

const FONT_STACK = {
  'default':        "'Malgun Gothic','맑은 고딕','Apple SD Gothic Neo',sans-serif",
  'noto-sans-kr':   "'Noto Sans KR',sans-serif",
  'gowun-dodum':    "'Gowun Dodum',sans-serif",
  'nanum-myeongjo': "'Nanum Myeongjo',serif",
  'black-han-sans': "'Black Han Sans',sans-serif",
};
const GOOGLE_FONTS = {
  'noto-sans-kr':   'https://fonts.googleapis.com/css2?family=Noto+Sans+KR:wght@400;500;700&display=swap',
  'gowun-dodum':    'https://fonts.googleapis.com/css2?family=Gowun+Dodum&display=swap',
  'nanum-myeongjo': 'https://fonts.googleapis.com/css2?family=Nanum+Myeongjo:wght@400;700&display=swap',
  'black-han-sans': 'https://fonts.googleapis.com/css2?family=Black+Han+Sans&display=swap',
};

function applyFont(fontKey) {
  const stack = FONT_STACK[fontKey] || FONT_STACK['default'];
  document.body.style.fontFamily = stack;
  if (GOOGLE_FONTS[fontKey]) {
    const link = document.createElement('link');
    link.rel = 'stylesheet'; link.href = GOOGLE_FONTS[fontKey];
    document.head.appendChild(link);
  }
}

function applyVideoBg(filename) {
  activeVideoBg = filename || '';
  const vid = document.getElementById('bg-video');
  const src = document.getElementById('bg-video-src');
  if (!vid || !src) return;
  if (!filename) {
    vid.style.display = 'none';
    if (!globalImageBgDisabled) {
      document.body.style.backgroundImage = "url('/display/bg')";
      document.body.style.backgroundSize = 'cover';
      document.body.style.backgroundPosition = 'center';
    } else {
      document.body.style.backgroundImage = 'none';
      document.body.style.background = '#000';
    }
    return;
  }
  src.src = '/display/video-bg/' + filename;
  vid.load();
  vid.style.display = 'block';
  document.body.style.backgroundImage = 'none';
  document.body.style.background = 'transparent';
}

async function initDisplayConfig() {
  try {
    const logoRes = await fetch('/api/logo', { method: 'HEAD' });
    if (logoRes.ok) logoUrl = '/api/logo';
  } catch (e) {}
  try {
    const cfgRes = await fetch('/api/display-config');
    if (cfgRes.ok) {
      const cfg = await cfgRes.json();
      applyFont(cfg.font);
      globalImageBgDisabled = !!cfg.globalImageBgDisabled;
      applyVideoBg(cfg.globalVideoBg || '');
      if (cfg.logoPosition) logoPosition = cfg.logoPosition;
      if (cfg.logoSizePercent) logoSizePercent = cfg.logoSizePercent;
    }
  } catch (e) {}
}

function paginate(text, linesPerPage) {
  const lines = text.split('\n').filter(l => l.trim() !== '');
  const pages = [];
  for (let i = 0; i < lines.length; i += linesPerPage)
    pages.push(lines.slice(i, i + linesPerPage).join('\n'));
  return pages.length ? pages : [text];
}

function esc(s) {
  return String(s)
    .replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;')
    .replace(/\n/g,'<br>');
}

function renderItem(item, pageIdx) {
  const info     = item.info     || '';
  const title    = item.title    || '';
  const obj      = item.obj      || '';
  const lead     = item.lead     || '';
  const contents = item.contents || '';
  const images   = item.images   || [];
  const bgImage  = item.bgImage  || '';

  if (bgImage) {
    slide.style.backgroundImage = "linear-gradient(rgba(0,0,0,0.35),rgba(0,0,0,0.35)), url('" + bgImage + "')";
  } else if (activeVideoBg) {
    slide.style.backgroundImage = "linear-gradient(rgba(0,0,0,0.35),rgba(0,0,0,0.35))";
  } else if (!globalImageBgDisabled) {
    slide.style.backgroundImage = "linear-gradient(rgba(0,0,0,0.4),rgba(0,0,0,0.4)), url('/display/bg')";
  } else {
    slide.style.backgroundImage = "linear-gradient(rgba(0,0,0,0.4),rgba(0,0,0,0.4))";
  }
  slide.className = 'visible';

  const posText  = slides.length ? (idx + 1) + ' / ' + slides.length : '';
  const pageText = subPages.length > 1 ? '1 / ' + subPages.length : '';
  const churchBox = logoUrl ? (function() {
    const vPos = logoPosition.startsWith('top') ? 'top:1.5vh' : 'bottom:1.5vh';
    const hPos = logoPosition.endsWith('right')  ? 'right:2vw'  : 'left:2vw';
    return '<div style="position:absolute;' + vPos + ';' + hPos + ';display:flex;align-items:flex-end">' +
      '<img src="' + logoUrl + '" alt="logo" style="max-height:7vh;max-width:' + logoSizePercent + 'vw;object-fit:contain;opacity:0.88;filter:drop-shadow(0 2px 6px rgba(0,0,0,0.55))"></div>';
  })() : '';
  const footer = '<div class="divider"></div>' +
    '<div class="slide-pos">' + posText + '</div>' +
    (pageText ? '<div class="page-indicator">' + pageText + '</div>' : '') + churchBox;
  const header =
    '<div class="label">' + esc(lead) + '</div>' +
    '<div class="order-title">' + esc(title) + '</div>';

  if (info.startsWith('b_') && contents) {
    const page = subPages[pageIdx] || contents;
    slide.innerHTML = header + '<div class="bible-ref">' + esc(obj) + '</div>' +
      '<div class="bible-contents">' + esc(page) + '</div>' + footer;
    return;
  }
  if (info === 'lyrics_display' && subPages.length > 0) {
    slide.innerHTML = header + '<div class="lyrics-text">' + esc(subPages[pageIdx] || '') + '</div>' + footer;
    return;
  }
  if (title === '찬송' || title === '헌금봉헌') {
    if (images.length > 0) {
      // 미리보기: 표지(숫자) 건너뛰고 첫 번째 이미지 페이지 직접 표시
      slide.style.backgroundImage = 'none';
      slide.innerHTML = '<img class="slide-image" src="' + images[0] + '">' + footer;
      return;
    }
    slide.innerHTML = header + '<div class="hymn-number">' + esc(obj) + '</div>' +
      (lead ? '<div class="hymn-sub">' + esc(lead) + '</div>' : '') + footer;
    return;
  }
  if (title === '성시교독') {
    if (images.length > 0) {
      slide.style.backgroundImage = 'none';
      slide.innerHTML = '<img class="slide-image" src="' + images[0] + '">' + footer;
      return;
    }
    slide.innerHTML = header + '<div class="hymn-number">' + esc(obj) + '</div>' +
      (lead ? '<div class="hymn-sub">' + esc(lead) + '</div>' : '') + footer;
    return;
  }
  if (title === '대표기도') {
    slide.innerHTML = header + '<div class="title">' + esc(title) + '</div>' +
      '<div class="prayer-name">' + esc(lead) + '</div>' + footer;
    return;
  }
  if (title === '신앙고백' && contents) {
    const page = subPages.length > 0 ? (subPages[pageIdx] || contents) : contents;
    slide.innerHTML = header + '<div class="title">' + esc(obj) + '</div>' +
      '<div class="creed-text">' + esc(page) + '</div>' + footer;
    return;
  }
  if (title === '주기도문' && contents) {
    const page = subPages.length > 0 ? (subPages[pageIdx] || contents) : contents;
    slide.innerHTML = header + '<div class="title">' + esc(title) + '</div>' +
      '<div class="creed-text">' + esc(page) + '</div>' + footer;
    return;
  }
  if (title === '참회의 기도') {
    slide.innerHTML = header + (bgImage ? '' : '<div class="title">' + esc(title) + '</div>') +
      '<div class="confession-text">' + esc(obj !== '-' ? obj : '') + '</div>' + footer;
    return;
  }
  if (title === '말씀') {
    slide.innerHTML = header + '<div class="sermon-title">' + esc(obj !== '-' ? obj : '') + '</div>' +
      (lead ? '<div class="sermon-pastor">' + esc(lead) + '</div>' : '') + footer;
    return;
  }
  if (info === 'notice' || title === '교회소식') {
    slide.innerHTML = header + '<div class="notice-title">' + esc(title) + '</div>' +
      '<div class="notice-contents">' + esc(contents || obj) + '</div>' + footer;
    return;
  }
  if (bgImage) {
    slide.style.backgroundImage = "url('" + bgImage + "')";
    slide.innerHTML = header + footer;
    return;
  }
  const mainText = (obj && obj !== '-') ? obj : '';
  slide.innerHTML = header + '<div class="title">' + esc(title) + '</div>' +
    (mainText ? '<div class="obj">' + esc(mainText) + '</div>' : '') +
    (lead && !mainText ? '<div class="prayer-name">' + esc(lead) + '</div>' : '') + footer;
}

initDisplayConfig().finally(function() {
  if (!slides.length) return;
  var item = slides[idx] || {};

  // pdf_only 항목: 미리보기에서는 "Display 건너뜀" 안내 표시
  if ((item.info || '') === 'pdf_only') {
    document.body.style.background = '#111';
    slide.style.backgroundImage = 'none';
    slide.className = 'visible';
    slide.innerHTML =
      '<div style="position:absolute;inset:0;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:2.5vh">' +
        '<div style="font-size:3vh;color:rgba(255,255,255,0.45)">' + esc(item.title || '') + '</div>' +
        '<div style="background:rgba(245,158,11,0.15);border:1px solid rgba(245,158,11,0.45);color:#f59e0b;padding:1vh 2vw;border-radius:0.5vw;font-size:2.2vh;letter-spacing:0.04em;">PDF 전용 — Display 건너뜀</div>' +
        (item.obj && item.obj !== '-' ? '<div style="font-size:2vh;color:rgba(255,255,255,0.3)">' + esc(item.obj) + '</div>' : '') +
      '</div>';
    return;
  }

  subPages = [];
  var itemTitle = item.title || '';
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
  renderItem(item, 0);
});
</script>
</body>
</html>`
