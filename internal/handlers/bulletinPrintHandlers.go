package handlers

import (
	"context"
	_ "embed"
	"easyPreparation_1.0/internal/bulletin/templates"
	"easyPreparation_1.0/internal/path"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"os/exec"
)

//go:embed html/bulletin-print.html
var bulletinPrintHTMLTmpl string

// ──────────────────── 데이터 타입 ────────────────────

// bulletinWorshipItem — config/xxx_worship.json 항목 구조
type bulletinWorshipItem struct {
	Title    string                `json:"title"`
	Obj      string                `json:"obj"`
	Lead     string                `json:"lead"`
	Info     string                `json:"info"`
	Key      string                `json:"key"`
	Contents string                `json:"contents,omitempty"`
	Children []bulletinWorshipItem `json:"children,omitempty"`
}

// BulletinOrderItem — 예배 순서 항목 (JSX 템플릿용)
type BulletinOrderItem struct {
	Item      string `json:"item"`
	Content   string `json:"content,omitempty"`
	Who       string `json:"who,omitempty"`
	Emphasize bool   `json:"emphasize,omitempty"`
}

// BulletinAnn — 교회 광고/소식
type BulletinAnn struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Fixed bool   `json:"fixed,omitempty"` // true: 고정 소식 (매주 동일)
	Order int    `json:"order,omitempty"` // 고정 소식 정렬 순서
}

// BulletinVerse — 성구
type BulletinVerse struct {
	Ref  string `json:"ref"`
	Text string `json:"text"`
}

// BulletinHelper — 봉사자
type BulletinHelper struct {
	Role string `json:"role"`
	Name string `json:"name"`
}

// BulletinData — JSX 템플릿에 전달하는 주보 데이터
type BulletinData struct {
	ChurchName    string              `json:"churchName"`
	ChurchNameEn  string              `json:"churchNameEn"`
	Pastor        string              `json:"pastor"`
	Website       string              `json:"website"`    // 교회 웹사이트 URL
	BlogInfo      string              `json:"blogInfo"`   // 블로그/SNS 표시 텍스트
	Tagline       string              `json:"tagline"`    // 교회 표어/모토
	Date          string              `json:"date"`
	WeekNumber    string              `json:"weekNumber"`
	Scripture     string              `json:"scripture"`
	SermonTitle   string              `json:"sermonTitle"`
	Order         []BulletinOrderItem `json:"order"`
	Afternoon     []BulletinOrderItem `json:"afternoon"`
	Wednesday     []BulletinOrderItem `json:"wednesday"`
	Announcements []BulletinAnn       `json:"announcements"`
	VerseQuote    BulletinVerse       `json:"verseQuote"`
	ClosingVerse  BulletinVerse       `json:"closingVerse"`
	Helpers       []BulletinHelper    `json:"helpers"`
	HymnNumbers   []string            `json:"hymnNumbers"`
	CoverImage    string              `json:"coverImage,omitempty"`
	AccentColor   string              `json:"accentColor,omitempty"` // 표지 이미지에서 추출한 주조색
}

// BulletinTheme — data/bulletin_theme.json 저장 구조
type BulletinTheme struct {
	AccentColor string `json:"accentColor"`
}

// ChurchInfo — 교회 기본 정보 (data/church_info.json)
type ChurchInfo struct {
	ChurchName   string `json:"churchName"`
	ChurchNameEn string `json:"churchNameEn"`
	Pastor       string `json:"pastor"`
	Website      string `json:"website"`  // 교회 웹사이트 URL
	BlogInfo     string `json:"blogInfo"` // 블로그/SNS 표시 텍스트
	Tagline      string `json:"tagline"`  // 교회 표어/모토
}

// ──────────────────── 템플릿 파일 맵 ────────────────────

var bulletinTemplateFiles = map[string]string{
	"1":  "v1-classic",
	"2":  "v2-minimal",
	"3":  "v3-editorial",
	"4":  "v4-nature",
	"5":  "v5-youth",
	"6":  "v6-korean-classic",
	"9":  "v9-timeline",
	"10": "v10-photo-pane",
}

// ──────────────────── HTML 호스트 페이지 ────────────────────

// BulletinPrintHandler — GET /display/bulletin-print?type=X&template=N
func BulletinPrintHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	worshipType := r.URL.Query().Get("type")
	if !validWorshipTypes[worshipType] {
		http.Error(w, "Invalid worship type", http.StatusBadRequest)
		return
	}
	templateNum := r.URL.Query().Get("template")
	tmplFile, ok := bulletinTemplateFiles[templateNum]
	if !ok {
		templateNum = "1"
		tmplFile = "v1-classic"
	}

	html := fmt.Sprintf(bulletinPrintHTMLTmpl,
		worshipType,     // 1: bulletin-data type (먼저 fetch)
		tmplFile+".jsx", // 2: JSX 파일명
		templateNum,     // 3: V{N}Outside
		templateNum,     // 4: V{N}Inside
		templateNum,     // 5: error msg
		templateNum,     // 6: error msg
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

// bulletinPrintHTMLTmpl is loaded via //go:embed at the top of this file
// Sprintf 인수: worshipType, jsxFile, tplNum(x4)

// ──────────────────── JSX 파일 서빙 ────────────────────

// BulletinTemplateFileHandler — GET /display/bulletin-template/{file}
func BulletinTemplateFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	fileName := filepath.Base(strings.TrimPrefix(r.URL.Path, "/display/bulletin-template/"))
	data, err := templates.FS.ReadFile(fileName)
	if err != nil {
		http.Error(w, "Template not found: "+fileName, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(data)
}

// ──────────────────── 주보 데이터 API ────────────────────

// BulletinDataHandler — GET /api/bulletin-data?type=X
func BulletinDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	worshipType := r.URL.Query().Get("type")
	if !validWorshipTypes[worshipType] {
		http.Error(w, "Invalid worship type", http.StatusBadRequest)
		return
	}
	bd, err := buildBulletinData(worshipType)
	if err != nil {
		http.Error(w, "데이터 로드 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(bd)
}

// ──────────────────── Chrome PDF 생성 ────────────────────

// BulletinPreviewHandler — GET /api/bulletin-preview?type=X&template=N
// Desktop 모드: OS 시스템 브라우저로 미리보기 열기 (Wails WebView window.open 차단 우회)
// 브라우저 모드: 403 반환 → 클라이언트가 window.open fallback
func BulletinPreviewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if desktopDownloadDir == "" {
		http.Error(w, "not desktop mode", http.StatusForbidden)
		return
	}
	worshipType := r.URL.Query().Get("type")
	if !validWorshipTypes[worshipType] {
		http.Error(w, "Invalid worship type", http.StatusBadRequest)
		return
	}
	templateNum := r.URL.Query().Get("template")
	if _, ok := bulletinTemplateFiles[templateNum]; !ok {
		templateNum = "1"
	}

	previewURL := fmt.Sprintf("http://localhost:8080/display/bulletin-print?type=%s&template=%s",
		worshipType, templateNum)
	openInBrowser(w, previewURL)
}

// BulletinPdfSaveHandler — GET /api/bulletin-pdf-save?type=X&template=N
// Desktop 모드 전용: PDF를 desktopDownloadDir에 저장 + 폴더 열기
// 데스크탑 모드가 아닐 때 403 반환 (클라이언트 fallback 신호)
func BulletinPdfSaveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if desktopDownloadDir == "" {
		http.Error(w, "not desktop mode", http.StatusForbidden)
		return
	}
	worshipType := r.URL.Query().Get("type")
	if !validWorshipTypes[worshipType] {
		http.Error(w, "Invalid worship type", http.StatusBadRequest)
		return
	}
	templateNum := r.URL.Query().Get("template")
	if _, ok := bulletinTemplateFiles[templateNum]; !ok {
		templateNum = "1"
	}

	pdfBytes, err := generateBulletinPDF(worshipType, templateNum)
	if err != nil {
		http.Error(w, "PDF 생성 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fileName := fmt.Sprintf("bulletin_%s_%s.pdf", worshipType, time.Now().Format("20060102"))
	savePath := filepath.Join(desktopDownloadDir, fileName)
	if err := os.WriteFile(savePath, pdfBytes, 0644); err != nil {
		http.Error(w, "저장 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	openFolder(desktopDownloadDir)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// BulletinPdfHandler — GET /api/bulletin-pdf?type=X&template=N
func BulletinPdfHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	worshipType := r.URL.Query().Get("type")
	if !validWorshipTypes[worshipType] {
		http.Error(w, "Invalid worship type", http.StatusBadRequest)
		return
	}
	templateNum := r.URL.Query().Get("template")
	if _, ok := bulletinTemplateFiles[templateNum]; !ok {
		templateNum = "1"
	}

	pdfBytes, err := generateBulletinPDF(worshipType, templateNum)
	if err != nil {
		http.Error(w, "PDF 생성 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fileName := fmt.Sprintf("bulletin_%s_%s.pdf", worshipType, time.Now().Format("20060102"))
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
	w.Write(pdfBytes)
}

// findChromeBin — 시스템에 설치된 Chrome/Chromium 경로 반환 (없으면 "")
func findChromeBin() string {
	candidates := []string{
		// Windows
		`C:\Program Files\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		filepath.Join(os.Getenv("LOCALAPPDATA"), `Google\Chrome\Application\chrome.exe`),
		// macOS
		`/Applications/Google Chrome.app/Contents/MacOS/Google Chrome`,
		`/Applications/Chromium.app/Contents/MacOS/Chromium`,
		// Linux
		`/usr/bin/google-chrome`,
		`/usr/bin/google-chrome-stable`,
		`/usr/bin/chromium-browser`,
		`/usr/bin/chromium`,
	}
	for _, p := range candidates {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// generateBulletinPDF — Chrome --print-to-pdf 로 주보 PDF 생성
func generateBulletinPDF(worshipType, templateNum string) ([]byte, error) {
	chromeBin := findChromeBin()
	if chromeBin == "" {
		return nil, fmt.Errorf("Chrome 또는 Chromium이 설치되어 있지 않습니다. Chrome을 설치한 후 다시 시도하세요.")
	}

	targetURL := fmt.Sprintf("http://localhost:8080/display/bulletin-print?type=%s&template=%s",
		worshipType, templateNum)
	outFile := filepath.Join(os.TempDir(), fmt.Sprintf("bulletin_%d.pdf", time.Now().UnixNano()))
	defer os.Remove(outFile)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, chromeBin,
		"--headless",
		"--disable-gpu",
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"--run-all-compositor-stages-before-draw",
		"--virtual-time-budget=15000",
		"--print-to-pdf-no-header",
		"--print-to-pdf="+outFile,
		targetURL,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("PDF 생성 실패: %w\n%s", err, string(out))
	}
	return os.ReadFile(outFile)
}

// ──────────────────── 주보 표지 이미지 API ────────────────────

// BulletinCoverHandler — GET/POST /api/bulletin-cover
// 주보 템플릿 커버에 사용되는 주간 표지 이미지 (매주 교체 가능)
func BulletinCoverHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	switch r.Method {
	case http.MethodGet:
		p := findBulletinCoverPath()
		if p == "" {
			http.NotFound(w, r)
			return
		}
		// 캐시 방지 (매주 교체 가능)
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		http.ServeFile(w, r, p)
	case http.MethodPost:
		if err := r.ParseMultipartForm(20 << 20); err != nil {
			http.Error(w, "파일 파싱 실패", http.StatusBadRequest)
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "파일 없음", http.StatusBadRequest)
			return
		}
		defer file.Close()
		ext := strings.ToLower(filepath.Ext(header.Filename))
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
			http.Error(w, "PNG/JPG만 허용합니다", http.StatusBadRequest)
			return
		}
		dataDir := filepath.Join(path.ExecutePath("easyPreparation"), "data")
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			http.Error(w, "디렉토리 생성 실패", http.StatusInternalServerError)
			return
		}
		// 기존 파일 삭제 (다른 확장자일 수 있으므로)
		for _, oldExt := range []string{".png", ".jpg", ".jpeg"} {
			_ = os.Remove(filepath.Join(dataDir, "bulletin_cover"+oldExt))
		}
		savePath := filepath.Join(dataDir, "bulletin_cover"+ext)
		dst, err := os.Create(savePath)
		if err != nil {
			http.Error(w, "파일 저장 실패", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "파일 쓰기 실패", http.StatusInternalServerError)
			return
		}
		// 표지 이미지에서 주조색 자동 추출 → bulletin_theme.json 저장
		if accent := extractDominantAccent(savePath); accent != "" {
			_ = saveBulletinTheme(BulletinTheme{AccentColor: accent})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func findBulletinCoverPath() string {
	dataDir := path.ExecutePath("easyPreparation")
	for _, ext := range []string{".png", ".jpg", ".jpeg"} {
		p := filepath.Join(dataDir, "data", "bulletin_cover"+ext)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// ──────────────────── 교회 정보 API ────────────────────

// ChurchInfoHandler — GET/PUT /api/church-info
func ChurchInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	switch r.Method {
	case http.MethodGet:
		ci := loadChurchInfo()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ci)
	case http.MethodPut, http.MethodPost:
		var ci ChurchInfo
		if err := json.NewDecoder(r.Body).Decode(&ci); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if err := saveChurchInfo(ci); err != nil {
			http.Error(w, "저장 실패: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// ──────────────────── 내부 헬퍼 ────────────────────

func churchInfoPath() string {
	return filepath.Join(path.ExecutePath("easyPreparation"), "data", "church_info.json")
}

func loadChurchInfo() ChurchInfo {
	if data, err := os.ReadFile(churchInfoPath()); err == nil {
		var ci ChurchInfo
		if json.Unmarshal(data, &ci) == nil && ci.ChurchName != "" {
			return ci
		}
	}
	// DB fallback
	if apiDB != nil {
		var name, englishName string
		row := apiDB.QueryRow("SELECT name, english_name FROM churches WHERE id=1 LIMIT 1")
		if row.Scan(&name, &englishName) == nil {
			return ChurchInfo{ChurchName: name, ChurchNameEn: englishName}
		}
	}
	return ChurchInfo{ChurchName: "교회", ChurchNameEn: "CHURCH"}
}

// loadFixedAnnouncements — data/fixed_announcements.json 에서 고정 소식 로드
func loadFixedAnnouncements() []BulletinAnn {
	p := filepath.Join(path.ExecutePath("easyPreparation"), "data", "fixed_announcements.json")
	data, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	var items []BulletinAnn
	if err := json.Unmarshal(data, &items); err != nil {
		return nil
	}
	return items
}

func saveChurchInfo(ci ChurchInfo) error {
	dirPath := filepath.Dir(churchInfoPath())
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(ci, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(churchInfoPath(), data, 0644)
}

func loadWorshipItemsForBulletin(worshipType string) ([]bulletinWorshipItem, error) {
	execPath := path.ExecutePath("easyPreparation")
	filePath := filepath.Join(execPath, "config", worshipType+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var items []bulletinWorshipItem
	return items, json.Unmarshal(data, &items)
}

func buildBulletinData(worshipType string) (BulletinData, error) {
	items, err := loadWorshipItemsForBulletin(worshipType)
	if err != nil {
		return BulletinData{}, fmt.Errorf("예배 순서 로드 실패 (%s): %w", worshipType, err)
	}

	var bd BulletinData
	ci := loadChurchInfo()
	bd.ChurchName = ci.ChurchName
	bd.ChurchNameEn = ci.ChurchNameEn
	bd.Pastor = ci.Pastor
	bd.Website = ci.Website
	bd.BlogInfo = ci.BlogInfo
	bd.Tagline = ci.Tagline

	now := time.Now()
	bd.Date = fmt.Sprintf("%d. %02d. %02d", now.Year(), now.Month(), now.Day())
	_, week := now.ISOWeek()
	bd.WeekNumber = fmt.Sprintf("제 %d 주차", week)

	bd.Order, bd.Announcements, bd.Helpers,
		bd.SermonTitle, bd.Scripture,
		bd.VerseQuote, bd.ClosingVerse,
		bd.HymnNumbers = mapWorshipItemsForBulletin(items)

	if worshipType == "main_worship" {
		if afterItems, err := loadWorshipItemsForBulletin("after_worship"); err == nil {
			bd.Afternoon = mapSimpleBulletinOrder(afterItems)
		}
		if wedItems, err := loadWorshipItemsForBulletin("wed_worship"); err == nil {
			bd.Wednesday = mapSimpleBulletinOrder(wedItems)
		}
	}

	// 표지 이미지: bulletin_cover 우선 → logo fallback
	if findBulletinCoverPath() != "" {
		bd.CoverImage = "/api/bulletin-cover"
	} else if findLogoPath() != "" {
		bd.CoverImage = "/api/logo"
	}

	// 주조색: bulletin_theme.json에서 로드
	if theme := loadBulletinTheme(); theme.AccentColor != "" {
		bd.AccentColor = theme.AccentColor
	}

	// 고정 소식 병합 (data/fixed_announcements.json)
	fixedAnns := loadFixedAnnouncements()
	if len(fixedAnns) > 0 {
		bd.Announcements = append(bd.Announcements, fixedAnns...)
	}

	// nil 슬라이스 → 빈 배열 (JSON null 방지)
	if bd.Order == nil {
		bd.Order = []BulletinOrderItem{}
	}
	if bd.Afternoon == nil {
		bd.Afternoon = []BulletinOrderItem{}
	}
	if bd.Wednesday == nil {
		bd.Wednesday = []BulletinOrderItem{}
	}
	if bd.Announcements == nil {
		bd.Announcements = []BulletinAnn{}
	}
	if bd.Helpers == nil {
		bd.Helpers = []BulletinHelper{}
	}
	if bd.HymnNumbers == nil {
		bd.HymnNumbers = []string{}
	}

	return bd, nil
}

// mapWorshipItemsForBulletin — 예배 순서 항목 분류
func mapWorshipItemsForBulletin(items []bulletinWorshipItem) (
	order []BulletinOrderItem,
	announcements []BulletinAnn,
	helpers []BulletinHelper,
	sermonTitle, scripture string,
	verseQuote, closingVerse BulletinVerse,
	hymnNumbers []string,
) {
	helperTitles := map[string]bool{
		"내주기도": true, "헌금, 안내": true, "헌금봉헌안내": true,
	}
	// content도 who도 없는 중복 순서 항목 필터용 (e.g. 프로젝터 전용 슬라이드가 두 번 들어온 경우)
	seenTitles := map[string]bool{}

	for _, item := range items {
		title := strings.TrimSpace(item.Title)
		obj := strings.TrimSpace(item.Obj)
		lead := strings.TrimSpace(item.Lead)

		if title == "말씀내용" {
			continue
		}

		// 교회소식
		if title == "교회소식" || item.Info == "notice" {
			announcements = collectBulletinAnnouncements(item.Children)
			continue
		}

		// 봉사자
		if helperTitles[title] {
			if obj != "" && obj != "-" {
				helpers = append(helpers, BulletinHelper{Role: title, Name: obj})
			}
			continue
		}

		// 오늘의 말씀 → closingVerse
		if title == "오늘의 말씀" {
			closingVerse = BulletinVerse{
				Ref:  obj,
				Text: bulletinStripVersePrefix(item.Contents),
			}
			if verseQuote.Text == "" {
				verseQuote = closingVerse
			}
			continue
		}

		// 일반 예배 순서
		content := obj
		if content == "-" || content == title {
			// obj가 title과 같으면 중복 (e.g. "전주" / "전주")
			content = ""
		}
		if strings.Contains(content, "\n") {
			// 줄바꿈 있는 텍스트는 프로젝터 내부용 — 주보에서 제외
			content = ""
		}
		who := lead
		if who == "-" {
			who = ""
		}

		emphasize := title == "말씀"

		// content도 who도 없는데 같은 title이 이미 나왔으면 → 프로젝터 전용 중복 슬라이드, 건너뜀
		if content == "" && who == "" && seenTitles[title] {
			continue
		}
		seenTitles[title] = true

		order = append(order, BulletinOrderItem{
			Item:      title,
			Content:   content,
			Who:       who,
			Emphasize: emphasize,
		})

		if title == "말씀" && content != "" {
			sermonTitle = content
		}
		if item.Info == "b_edit" && title == "성경봉독" && obj != "" && obj != "-" {
			scripture = obj
		}
		if item.Info == "c_edit" && (title == "찬송" || title == "헌금봉헌") && obj != "" && obj != "-" {
			hymnNumbers = append(hymnNumbers, obj)
		}
	}

	// verseQuote fallback: 성경봉독 첫 구절
	if verseQuote.Text == "" {
		for _, item := range items {
			if item.Title == "성경봉독" && item.Contents != "" {
				verseQuote = BulletinVerse{
					Ref:  scripture,
					Text: bulletinExtractFirstVerse(item.Contents),
				}
				break
			}
		}
	}
	return
}

// mapSimpleBulletinOrder — 오후/수요 예배 단순 매핑
func mapSimpleBulletinOrder(items []bulletinWorshipItem) []BulletinOrderItem {
	var result []BulletinOrderItem
	for _, item := range items {
		t := strings.TrimSpace(item.Title)
		if t == "" || t == "-" {
			continue
		}
		c := strings.TrimSpace(item.Obj)
		if c == "-" {
			c = ""
		}
		result = append(result, BulletinOrderItem{Item: t, Content: c})
	}
	return result
}

// collectBulletinAnnouncements — 교회소식 children → 광고 목록 (재귀)
func collectBulletinAnnouncements(items []bulletinWorshipItem) []BulletinAnn {
	var anns []BulletinAnn
	for _, item := range items {
		if len(item.Children) > 0 {
			anns = append(anns, collectBulletinAnnouncements(item.Children)...)
		} else {
			t := strings.TrimSpace(item.Title)
			b := strings.TrimSpace(item.Obj)
			if t != "" && t != "-" {
				anns = append(anns, BulletinAnn{Title: t, Body: b})
			}
		}
	}
	return anns
}

var bulletinVersePrefixRe = regexp.MustCompile(`^\d+:\d+\s*[○●]?\s*`)

// bulletinStripVersePrefix — "62:7 본문 내용" → "본문 내용" (멀티라인 지원)
func bulletinStripVersePrefix(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	var parts []string
	for _, line := range lines {
		stripped := bulletinVersePrefixRe.ReplaceAllString(strings.TrimSpace(line), "")
		if stripped != "" {
			parts = append(parts, stripped)
		}
	}
	return strings.Join(parts, " ")
}

// bulletinExtractFirstVerse — 멀티라인 성구에서 첫 번째 구절 텍스트 추출
func bulletinExtractFirstVerse(contents string) string {
	for _, line := range strings.Split(strings.TrimSpace(contents), "\n") {
		stripped := bulletinVersePrefixRe.ReplaceAllString(strings.TrimSpace(line), "")
		if stripped != "" {
			return stripped
		}
	}
	return ""
}

// ──────────────────── 주보 테마(주조색) API ────────────────────

// BulletinThemeHandler — GET/PUT /api/bulletin-theme
func BulletinThemeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(loadBulletinTheme())
	case http.MethodPut, http.MethodPost:
		var t BulletinTheme
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if err := saveBulletinTheme(t); err != nil {
			http.Error(w, "저장 실패: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func bulletinThemePath() string {
	return filepath.Join(path.ExecutePath("easyPreparation"), "data", "bulletin_theme.json")
}

func loadBulletinTheme() BulletinTheme {
	data, err := os.ReadFile(bulletinThemePath())
	if err != nil {
		return BulletinTheme{}
	}
	var t BulletinTheme
	json.Unmarshal(data, &t)
	return t
}

func saveBulletinTheme(t BulletinTheme) error {
	dirPath := filepath.Dir(bulletinThemePath())
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(bulletinThemePath(), data, 0644)
}

// ──────────────────── 표지 이미지 주조색 추출 ────────────────────

// extractDominantAccent — 이미지 파일에서 주조색 추출 (표지 이미지 → 시안 액센트 색)
// 전략: 채도 높고 밝기가 중간인 픽셀만 샘플링 → 30° 단위 12개 색조 버킷 → 가장 많은 버킷 평균 → 텍스트용으로 어둡게
func extractDominantAccent(imgPath string) string {
	f, err := os.Open(imgPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return ""
	}

	bounds := img.Bounds()
	var colorSums [12][3]float64
	var counts [12]int

	// 4픽셀 간격으로 샘플링 (속도 최적화)
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 4 {
		for x := bounds.Min.X; x < bounds.Max.X; x += 4 {
			c := img.At(x, y)
			r32, g32, b32, a32 := c.RGBA()
			if a32 < 32768 {
				continue // 반투명 픽셀 제외
			}
			r := float64(r32) / 65535.0
			g := float64(g32) / 65535.0
			b := float64(b32) / 65535.0

			h, s, l := rgbToHsl(r, g, b)
			// 무채색, 매우 밝은, 매우 어두운 픽셀 제외
			if s < 0.18 || l < 0.12 || l > 0.88 {
				continue
			}

			bucket := int(h/30.0) % 12
			colorSums[bucket][0] += r
			colorSums[bucket][1] += g
			colorSums[bucket][2] += b
			counts[bucket]++
		}
	}

	// 가장 많은 버킷 찾기
	maxIdx := 0
	for i := 1; i < 12; i++ {
		if counts[i] > counts[maxIdx] {
			maxIdx = i
		}
	}
	if counts[maxIdx] == 0 {
		return ""
	}

	// 평균 색상 계산
	n := float64(counts[maxIdx])
	avgR := colorSums[maxIdx][0] / n
	avgG := colorSums[maxIdx][1] / n
	avgB := colorSums[maxIdx][2] / n

	// 텍스트/액센트 용도: 채도 높이고 밝기 낮춤
	h, s, l := rgbToHsl(avgR, avgG, avgB)
	if s < 0.45 {
		s = 0.45
	}
	if l > 0.38 {
		l = 0.32
	} else if l < 0.18 {
		l = 0.22
	}

	outR, outG, outB := hslToRgb(h, s, l)
	return fmt.Sprintf("#%02x%02x%02x",
		int(math.Round(outR*255)),
		int(math.Round(outG*255)),
		int(math.Round(outB*255)),
	)
}

// rgbToHsl — RGB → HSL 변환 (모두 0.0~1.0 범위)
func rgbToHsl(r, g, b float64) (h, s, l float64) {
	maxC := math.Max(r, math.Max(g, b))
	minC := math.Min(r, math.Min(g, b))
	l = (maxC + minC) / 2.0

	if maxC == minC {
		return 0, 0, l
	}

	d := maxC - minC
	if l > 0.5 {
		s = d / (2.0 - maxC - minC)
	} else {
		s = d / (maxC + minC)
	}

	switch maxC {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	case b:
		h = (r-g)/d + 4
	}
	h *= 60
	return
}

func hslHue(p, q, t float64) float64 {
	if t < 0 {
		t++
	}
	if t > 1 {
		t--
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

// hslToRgb — HSL → RGB 변환 (h: 0~360, s/l: 0.0~1.0 → r/g/b: 0.0~1.0)
func hslToRgb(h, s, l float64) (r, g, b float64) {
	if s == 0 {
		return l, l, l
	}
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	hn := h / 360.0
	r = hslHue(p, q, hn+1.0/3.0)
	g = hslHue(p, q, hn)
	b = hslHue(p, q, hn-1.0/3.0)
	return
}
