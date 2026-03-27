package handlers

import (
	"easyPreparation_1.0/internal/quote"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const displayHTML = `<!DOCTYPE html>
<html lang="ko">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Display</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }

  html, body {
    width: 100%;
    height: 100%;
    background: #000;
    color: #fff;
    font-family: 'Malgun Gothic', '맑은 고딕', 'Apple SD Gothic Neo', sans-serif;
    overflow: hidden;
  }

  #slide {
    width: 100%;
    height: 100vh;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    padding: 7.4vh 6.25vw;
    opacity: 0;
    transition: opacity 0.5s ease;
    position: relative;
  }
  #slide.visible { opacity: 1; }

  #slide.empty { background: #000; }

  /* 상단 레이블 (인도자) */
  .label {
    position: absolute;
    top: 4.4vh;
    left: 4.2vw;
    font-size: 2.6vh;
    color: rgba(255,255,255,0.45);
    letter-spacing: 0.05em;
  }

  /* 우상단 순서 제목 */
  .order-title {
    position: absolute;
    top: 4.4vh;
    right: 4.2vw;
    font-size: 2.6vh;
    color: rgba(255,255,255,0.45);
    letter-spacing: 0.05em;
  }

  /* 메인 제목 */
  .title {
    font-size: 5.9vh;
    font-weight: 700;
    text-align: center;
    margin-bottom: 3.3vh;
    line-height: 1.3;
    letter-spacing: -0.01em;
  }

  /* 부제 / 찬송 번호 */
  .obj {
    font-size: 4.4vh;
    font-weight: 400;
    color: rgba(255,255,255,0.75);
    text-align: center;
    line-height: 1.5;
  }

  /* 성경 본문 */
  .bible-contents {
    font-size: 3.3vh;
    line-height: 1.9;
    text-align: left;
    color: #fff;
    max-height: 74vh;
    overflow: hidden;
    white-space: pre-wrap;
    width: 100%;
  }

  .bible-ref {
    font-size: 3vh;
    color: rgba(255,255,255,0.6);
    margin-bottom: 3.7vh;
    text-align: left;
    width: 100%;
  }

  /* 광고/공지 */
  .notice-title {
    font-size: 4.8vh;
    font-weight: 700;
    margin-bottom: 2.6vh;
    color: #fff;
  }
  .notice-obj {
    font-size: 3.5vh;
    color: rgba(255,255,255,0.8);
    text-align: center;
    line-height: 1.8;
    white-space: pre-wrap;
  }

  /* 하단 구분선 */
  .divider {
    position: absolute;
    bottom: 5.6vh;
    left: 4.2vw;
    right: 4.2vw;
    height: 2px;
    background: rgba(255,255,255,0.12);
  }
</style>
</head>
<body>
<div id="slide" class="empty"></div>

<script>
  const slide = document.getElementById('slide');
  let ws;
  let reconnectTimer;

  function connect() {
    const proto = location.protocol === 'https:' ? 'wss' : 'ws';
    ws = new WebSocket(proto + '://' + location.host + '/ws');

    ws.onmessage = (e) => {
      const msg = JSON.parse(e.data);
      if (msg.type === 'display') render(msg);
    };

    ws.onclose = () => {
      clearTimeout(reconnectTimer);
      reconnectTimer = setTimeout(connect, 3000);
    };
  }

  function render(data) {
    const info = data.info || '';
    const title = data.title || '';
    const obj   = data.obj   || '';
    const lead  = data.lead  || '';
    const contents = data.contents || '';

    slide.className = 'visible';

    // 성경 본문
    if (info === 'b_edit' && contents) {
      slide.innerHTML = ` + "`" + `
        <div class="label">${lead}</div>
        <div class="order-title">${title}</div>
        <div class="bible-ref">${obj}</div>
        <div class="bible-contents">${escHtml(contents)}</div>
        <div class="divider"></div>
      ` + "`" + `;
      return;
    }

    // 공지/광고
    if (info === 'notice' || info === 'c-edit') {
      slide.innerHTML = ` + "`" + `
        <div class="label">${lead}</div>
        <div class="notice-title">${title}</div>
        <div class="notice-obj">${escHtml(obj)}</div>
        <div class="divider"></div>
      ` + "`" + `;
      return;
    }

    // 기본 (찬송, 기도, 일반)
    slide.innerHTML = ` + "`" + `
      <div class="label">${lead}</div>
      <div class="order-title">${title}</div>
      <div class="title">${escHtml(obj || title)}</div>
      ${obj && obj !== '-' && obj !== title ? '<div class="obj">' + escHtml(obj) + '</div>' : ''}
      <div class="divider"></div>
    ` + "`" + `;
  }

  function escHtml(s) {
    return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;')
            .replace(/\n/g,'<br>');
  }

  connect();
</script>
</body>
</html>`

// DisplayHandler — GET /display: OBS Browser Source용 전체화면 슬라이드 페이지
func DisplayHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(displayHTML))
}

// DisplayPushHandler — POST /display/push: 운영자가 표시할 항목 전송
// b_edit 항목이고 contents가 없으면 DB에서 성경 본문 자동 조회
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

	// b_edit 항목이고 contents가 비어 있으면 DB에서 성경 본문 조회
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

// fetchBibleText — "책명_코드/장:절-절, ..." 형식에서 성경 본문 조회
// 반환: (본문텍스트, 한글구절표기)
func fetchBibleText(obj string) (string, string) {
	var texts []string
	var refs []string

	for _, part := range strings.Split(obj, ",") {
		part = strings.TrimSpace(part)
		// "신명기_5/4:5-4:6" 형태
		underIdx := strings.Index(part, "_")
		if underIdx < 0 {
			continue
		}
		korName := part[:underIdx]
		codeAndRange := part[underIdx+1:] // "5/4:5-4:6"

		// 책 번호 / 구절 범위 분리
		slashIdx := strings.Index(codeAndRange, "/")
		if slashIdx < 0 {
			continue
		}
		verseRange := codeAndRange[slashIdx+1:] // "4:5-4:6"

		text, err := quote.GetQuote(codeAndRange)
		if err != nil {
			continue
		}
		texts = append(texts, text)
		refs = append(refs, fmt.Sprintf("%s %s", korName, verseRange))
	}

	return strings.Join(texts, "\n"), strings.Join(refs, ", ")
}
