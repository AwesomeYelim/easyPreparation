package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"
)

// buildPrintLogoBox — 프린트 슬라이드용 로고 HTML 생성
func buildPrintLogoBox(logoExists bool, logoPosition string, logoSizePercent float64) string {
	if !logoExists {
		return ""
	}
	if logoPosition == "" {
		logoPosition = "bottom-right"
	}
	if logoSizePercent == 0 {
		logoSizePercent = 18
	}
	var vPos, hPos string
	if strings.HasPrefix(logoPosition, "top") {
		vPos = "top:1.5vh"
	} else {
		vPos = "bottom:1.5vh"
	}
	if strings.HasSuffix(logoPosition, "right") {
		hPos = "right:2vw"
	} else {
		hPos = "left:2vw"
	}
	return fmt.Sprintf(
		`<div style="position:absolute;%s;%s;display:flex;align-items:flex-end"><img src="/api/logo" alt="logo" style="max-height:7vh;max-width:%.0fvw;object-fit:contain;opacity:0.88;filter:drop-shadow(0 2px 6px rgba(0,0,0,0.55))"></div>`,
		vPos, hPos, logoSizePercent,
	)
}

// HandleDisplayPrint — /display/print
// 현재 예배 순서의 모든 슬라이드를 Display와 동일한 스타일로 렌더링한 인쇄용 HTML 반환.
// ?autoprint=1 쿼리 파라미터를 전달하면 페이지 로드 시 자동으로 인쇄 다이얼로그가 열림.
func HandleDisplayPrint(w http.ResponseWriter, r *http.Request) {
	autoprint := r.URL.Query().Get("autoprint") == "1"

	// 로고 + Display 설정 로드
	cfg := loadDisplayConfig()
	logoExists := findLogoPath() != ""
	logoBox := buildPrintLogoBox(logoExists, cfg.LogoPosition, cfg.LogoSizePercent)

	orderMu.RLock()
	items := deepCopyOrder(currentOrder)
	orderMu.RUnlock()

	// 예배 순서가 없으면 에러 페이지 반환
	if len(items) == 0 {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!DOCTYPE html><html lang="ko"><head><meta charset="UTF-8"><title>예배 PDF</title>
<style>*{margin:0;padding:0;box-sizing:border-box;}html,body{background:#000;color:#fff;font-family:'Malgun Gothic',sans-serif;height:100vh;display:flex;align-items:center;justify-content:center;flex-direction:column;gap:16px;}p{font-size:18px;opacity:.7;}small{font-size:13px;opacity:.4;}</style>
</head><body><p>예배 순서가 전송되지 않았습니다.</p><small>예배 PDF 버튼을 다시 클릭하거나, 먼저 "프로젝터에 보내기"를 실행하세요.</small></body></html>`))
		return
	}

	// 배경 이미지 URL (Display와 동일하게 서버 기본 배경 사용)
	const bgURL = "/display/bg"

	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html lang="ko">
<head>
<meta charset="UTF-8">
<title>예배 순서 — Print</title>
<style>
  @import url('https://fonts.googleapis.com/css2?family=Noto+Sans+KR:wght@400;500;700&display=swap');

  * { margin:0; padding:0; box-sizing:border-box; }

  html, body {
    background: #000;
    color: #fff;
    font-family: 'Malgun Gothic','맑은 고딕','Apple SD Gothic Neo',sans-serif;
  }

  .slide {
    width: 100vw;
    height: 100vh;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    padding: 7.4vh 6.25vw;
    position: relative;
    background-size: cover;
    background-position: center;
    background-repeat: no-repeat;
    page-break-after: always;
    break-after: page;
    overflow: hidden;
  }
  .slide:last-child { page-break-after: avoid; break-after: avoid; }

  /* 좌상단 인도자 */
  .label {
    position: absolute; top: 4.4vh; left: 4.2vw;
    font-size: 2.6vh; color: rgba(255,255,255,0.5);
    letter-spacing: 0.05em;
  }
  /* 우상단 순서 제목 */
  .order-title {
    position: absolute; top: 4.4vh; right: 4.2vw;
    font-size: 2.6vh; color: rgba(255,255,255,0.5);
  }
  /* 페이지 표시 */
  .page-indicator {
    position: absolute; bottom: 3vh; right: 4.2vw;
    font-size: 2vh; color: rgba(255,255,255,0.35);
  }
  /* 기본 제목 */
  .title {
    font-size: 5.9vh; font-weight: 700;
    text-align: center; margin-bottom: 3.3vh;
    line-height: 1.3; letter-spacing: -0.01em;
    text-shadow: 0 2px 8px rgba(0,0,0,0.6);
  }
  .obj {
    font-size: 4.4vh; font-weight: 400;
    color: rgba(255,255,255,0.85);
    text-align: center; line-height: 1.5;
    text-shadow: 0 2px 6px rgba(0,0,0,0.5);
  }
  /* 성경 본문 */
  .bible-ref {
    font-size: 3.5vh; color: rgba(255,255,255,0.65);
    margin-bottom: 3vh; text-align: left; width: 100%;
  }
  .bible-contents {
    font-size: 5vh; line-height: 1.9;
    text-align: left; color: #fff;
    white-space: pre-wrap; width: 100%;
  }
  /* 찬송/교독 */
  .hymn-number {
    font-size: 10vh; font-weight: 700;
    text-align: center; line-height: 1.2;
  }
  .hymn-sub {
    font-size: 3.2vh; color: rgba(255,255,255,0.6);
    margin-top: 2vh; text-align: center;
  }
  /* 이미지 슬라이드 */
  .slide-image {
    max-width: 90vw; max-height: 85vh;
    object-fit: contain;
  }
  /* 기도자 */
  .prayer-name {
    font-size: 8vh; font-weight: 700;
    text-align: center; line-height: 1.3;
  }
  /* 가사 슬라이드 */
  .lyrics-text {
    font-size: 7vh; line-height: 1.8;
    text-align: center; color: #fff;
    white-space: pre-wrap; width: 100%;
    font-weight: 500;
  }
  /* 사도신경/주기도문 */
  .creed-text {
    font-size: 3.3vh; line-height: 1.9;
    text-align: center; color: rgba(255,255,255,0.9);
    white-space: pre-wrap; width: 100%;
  }
  /* 참회의 기도 */
  .confession-text {
    font-size: 3.6vh; line-height: 2;
    text-align: left; color: rgba(255,255,255,0.9);
    white-space: pre-wrap; width: 100%;
  }
  /* 말씀 */
  .sermon-title {
    font-size: 7vh; font-weight: 700;
    text-align: center; line-height: 1.3;
    margin-bottom: 3vh;
  }
  .sermon-pastor {
    font-size: 4vh; color: rgba(255,255,255,0.7);
    text-align: center;
  }
  /* 교회소식 */
  .notice-title {
    font-size: 4.8vh; font-weight: 700;
    margin-bottom: 2.6vh; color: #fff;
  }
  .notice-contents {
    font-size: 2.8vh; color: rgba(255,255,255,0.85);
    text-align: left; line-height: 1.8;
    white-space: pre-wrap; width: 100%;
  }
  /* 하단 구분선 */
  .divider {
    position: absolute; bottom: 5.6vh; left: 4.2vw; right: 4.2vw;
    height: 1px; background: rgba(255,255,255,0.15);
  }
  .slide-pos {
    position: absolute; bottom: 3vh; left: 4.2vw;
    font-size: 2vh; color: rgba(255,255,255,0.35);
  }

  /* 인쇄 설정 */
  @media print {
    html, body { background: #000 !important; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
    .slide { width: 100%; height: 100vh; }
  }

  /* 버튼 (화면 전용) */
  .print-bar {
    position: fixed; top: 0; left: 0; right: 0; z-index: 9999;
    background: rgba(0,0,0,0.85); padding: 10px 20px;
    display: flex; align-items: center; gap: 12px;
    backdrop-filter: blur(6px);
  }
  .print-bar button {
    padding: 6px 18px; background: #3B82F6; color: #fff;
    border: none; border-radius: 8px; font-size: 13px; font-weight: 600;
    cursor: pointer;
  }
  .print-bar button:hover { background: #2563EB; }
  .print-bar .close-btn { background: #444; }
  .print-bar .close-btn:hover { background: #666; }
  .print-bar span { color: rgba(255,255,255,0.6); font-size: 12px; }
  @media print { .print-bar { display: none; } }
</style>
</head>
<body>
<div class="print-bar">
  <button onclick="window.print()">🖨 PDF로 저장 / 인쇄</button>
  <button class="close-btn" onclick="window.close()">닫기</button>
  <span>총 `)

	// 슬라이드 수 계산 (먼저 세고 나중에 렌더)
	totalSlides := countPrintSlides(items)
	sb.WriteString(fmt.Sprintf(`%d페이지`, totalSlides))
	sb.WriteString(` · Ctrl+P로 PDF 저장</span>
</div>
`)

	// 각 아이템 렌더링
	slideNum := 0
	for i, item := range items {
		title := strVal(item, "title")
		obj := strVal(item, "obj")
		lead := strVal(item, "lead")
		info := strVal(item, "info")
		contents := strVal(item, "contents")
		bgImage := strVal(item, "bgImage")

		posText := fmt.Sprintf("%d / %d", i+1, len(items))

		bgStyle := fmt.Sprintf(`background-image:linear-gradient(rgba(0,0,0,0.4),rgba(0,0,0,0.4)),url('%s')`, bgURL)
		if bgImage != "" {
			bgStyle = fmt.Sprintf(`background-image:linear-gradient(rgba(0,0,0,0.35),rgba(0,0,0,0.35)),url('%s')`, bgImage)
		}

		header := fmt.Sprintf(`<div class="label">%s</div><div class="order-title">%s</div>`,
			html.EscapeString(lead), html.EscapeString(title))
		footer := fmt.Sprintf(`<div class="divider"></div><div class="slide-pos">%s</div>%s`, posText, logoBox)

		// 이미지 슬라이드 — 배경 없이 이미지만
		images := getImageSlice(item, "images")

		// ── 성경 본문 ──
		if strings.HasPrefix(info, "b_") && contents != "" {
			pages := paginateText(contents, 3)
			for p, pg := range pages {
				pageText := ""
				if len(pages) > 1 {
					pageText = fmt.Sprintf(`<div class="page-indicator">%d / %d</div>`, p+1, len(pages))
				}
				sb.WriteString(fmt.Sprintf(`<div class="slide" style="%s">%s<div class="bible-ref">%s</div><div class="bible-contents">%s</div>%s%s</div>`,
					bgStyle, header,
					html.EscapeString(obj),
					html.EscapeString(pg),
					footer, pageText))
				slideNum++
			}
			continue
		}

		// ── 가사 슬라이드 ──
		if info == "lyrics_display" {
			pages := getLyricsPages(item)
			for p, pg := range pages {
				pageText := ""
				if len(pages) > 1 {
					pageText = fmt.Sprintf(`<div class="page-indicator">%d / %d</div>`, p+1, len(pages))
				}
				sb.WriteString(fmt.Sprintf(`<div class="slide" style="%s">%s<div class="lyrics-text">%s</div>%s%s</div>`,
					bgStyle, header,
					html.EscapeString(pg),
					footer, pageText))
				slideNum++
			}
			continue
		}

		// ── 찬송 / 헌금봉헌 (표지 + 이미지) ──
		if title == "찬송" || title == "헌금봉헌" {
			// 표지 슬라이드
			sb.WriteString(fmt.Sprintf(`<div class="slide" style="%s">%s<div class="hymn-number">%s</div>%s%s</div>`,
				bgStyle, header,
				html.EscapeString(obj),
				footer,
				func() string {
					if lead != "" {
						return fmt.Sprintf(`<div class="hymn-sub">%s</div>`, html.EscapeString(lead))
					}
					return ""
				}()))
			slideNum++
			// 이미지 페이지들
			for p, img := range images {
				pageText := fmt.Sprintf(`<div class="page-indicator">%d / %d</div>`, p+1, len(images))
				sb.WriteString(fmt.Sprintf(`<div class="slide" style="background:#000"><img class="slide-image" src="%s">%s%s</div>`,
					html.EscapeString(img), footer, pageText))
				slideNum++
			}
			continue
		}

		// ── 성시교독 (이미지만) ──
		if title == "성시교독" {
			if len(images) > 0 {
				for p, img := range images {
					pageText := ""
					if len(images) > 1 {
						pageText = fmt.Sprintf(`<div class="page-indicator">%d / %d</div>`, p+1, len(images))
					}
					sb.WriteString(fmt.Sprintf(`<div class="slide" style="background:#000"><img class="slide-image" src="%s">%s%s</div>`,
						html.EscapeString(img), footer, pageText))
					slideNum++
				}
				continue
			}
		}

		// ── 대표기도 ──
		if title == "대표기도" {
			sb.WriteString(fmt.Sprintf(`<div class="slide" style="%s">%s<div class="title">%s</div><div class="prayer-name">%s</div>%s</div>`,
				bgStyle, header,
				html.EscapeString(title),
				html.EscapeString(lead),
				footer))
			slideNum++
			continue
		}

		// ── 신앙고백 ──
		if title == "신앙고백" && contents != "" {
			pages := paginateText(contents, 10)
			for p, pg := range pages {
				pageText := ""
				if len(pages) > 1 {
					pageText = fmt.Sprintf(`<div class="page-indicator">%d / %d</div>`, p+1, len(pages))
				}
				sb.WriteString(fmt.Sprintf(`<div class="slide" style="%s">%s<div class="title">%s</div><div class="creed-text">%s</div>%s%s</div>`,
					bgStyle, header,
					html.EscapeString(obj),
					html.EscapeString(pg),
					footer, pageText))
				slideNum++
			}
			continue
		}

		// ── 주기도문 ──
		if title == "주기도문" && contents != "" {
			pages := paginateText(contents, 10)
			for p, pg := range pages {
				pageText := ""
				if len(pages) > 1 {
					pageText = fmt.Sprintf(`<div class="page-indicator">%d / %d</div>`, p+1, len(pages))
				}
				sb.WriteString(fmt.Sprintf(`<div class="slide" style="%s">%s<div class="title">%s</div><div class="creed-text">%s</div>%s%s</div>`,
					bgStyle, header,
					html.EscapeString(title),
					html.EscapeString(pg),
					footer, pageText))
				slideNum++
			}
			continue
		}

		// ── 참회의 기도 ──
		if title == "참회의 기도" {
			titleHTML := fmt.Sprintf(`<div class="title">%s</div>`, html.EscapeString(title))
			if bgImage != "" {
				titleHTML = ""
			}
			confText := obj
			if obj == "-" {
				confText = ""
			}
			sb.WriteString(fmt.Sprintf(`<div class="slide" style="%s">%s%s<div class="confession-text">%s</div>%s</div>`,
				bgStyle, header, titleHTML,
				html.EscapeString(confText),
				footer))
			slideNum++
			continue
		}

		// ── 말씀 ──
		if title == "말씀" {
			sermonText := obj
			if obj == "-" {
				sermonText = ""
			}
			sb.WriteString(fmt.Sprintf(`<div class="slide" style="%s">%s<div class="sermon-title">%s</div>%s%s</div>`,
				bgStyle, header,
				html.EscapeString(sermonText),
				func() string {
					if lead != "" {
						return fmt.Sprintf(`<div class="sermon-pastor">%s</div>`, html.EscapeString(lead))
					}
					return ""
				}(),
				footer))
			slideNum++
			continue
		}

		// ── 교회소식 ──
		if info == "notice" || title == "교회소식" {
			noticeText := contents
			if noticeText == "" {
				noticeText = obj
			}
			sb.WriteString(fmt.Sprintf(`<div class="slide" style="%s">%s<div class="notice-title">%s</div><div class="notice-contents">%s</div>%s</div>`,
				bgStyle, header,
				html.EscapeString(title),
				html.EscapeString(noticeText),
				footer))
			slideNum++
			continue
		}

		// ── 배경 이미지만 (bgImage 있는 단순 항목) ──
		if bgImage != "" {
			sb.WriteString(fmt.Sprintf(`<div class="slide" style="background-image:url('%s');background-size:cover;background-position:center">%s</div>`,
				html.EscapeString(bgImage), footer))
			slideNum++
			continue
		}

		// ── 기본 (제목 + obj) ──
		mainText := obj
		if obj == "-" {
			mainText = ""
		}
		objHTML := ""
		if mainText != "" {
			objHTML = fmt.Sprintf(`<div class="obj">%s</div>`, html.EscapeString(mainText))
		}
		prayerHTML := ""
		if lead != "" && mainText == "" {
			prayerHTML = fmt.Sprintf(`<div class="prayer-name">%s</div>`, html.EscapeString(lead))
		}
		sb.WriteString(fmt.Sprintf(`<div class="slide" style="%s">%s<div class="title">%s</div>%s%s%s</div>`,
			bgStyle, header,
			html.EscapeString(title),
			objHTML, prayerHTML,
			footer))
		slideNum++
	}

	if autoprint {
		sb.WriteString(`<script>window.addEventListener('load',function(){setTimeout(function(){window.print();},1000);});</script>`)
	}
	sb.WriteString(`</body></html>`)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	fmt.Fprint(w, sb.String())
}

// ─── 헬퍼 함수들 ───

func strVal(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getImageSlice(m map[string]interface{}, key string) []string {
	raw, ok := m[key]
	if !ok {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return v
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}

// getLyricsPages — item의 pages 필드를 []string으로 반환
func getLyricsPages(m map[string]interface{}) []string {
	raw, ok := m["pages"]
	if !ok {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return v
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}

// countPrintSlides — 총 슬라이드(페이지) 수를 미리 계산
func countPrintSlides(items []map[string]interface{}) int {
	count := 0
	for _, item := range items {
		title := strVal(item, "title")
		info := strVal(item, "info")
		contents := strVal(item, "contents")
		images := getImageSlice(item, "images")
		bgImage := strVal(item, "bgImage")

		switch {
		case strings.HasPrefix(info, "b_") && contents != "":
			count += len(paginateText(contents, 3))
		case info == "lyrics_display":
			pages := getLyricsPages(item)
			if len(pages) == 0 {
				count++
			} else {
				count += len(pages)
			}
		case title == "찬송" || title == "헌금봉헌":
			count += 1 + len(images)
		case title == "성시교독" && len(images) > 0:
			count += len(images)
		case title == "신앙고백" && contents != "":
			count += len(paginateText(contents, 10))
		case title == "주기도문" && contents != "":
			count += len(paginateText(contents, 10))
		case bgImage != "" && (title == "전주" || title == "마침"):
			count++
		default:
			count++
		}
	}
	return count
}

// HandleDisplayPrintJSON — /api/display/print-info (JSON, 슬라이드 수 등)
func HandleDisplayPrintJSON(w http.ResponseWriter, r *http.Request) {
	orderMu.RLock()
	items := deepCopyOrder(currentOrder)
	orderMu.RUnlock()

	total := countPrintSlides(items)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total":     total,
		"itemCount": len(items),
	})
}
