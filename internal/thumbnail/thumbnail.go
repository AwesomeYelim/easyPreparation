package thumbnail

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	xdraw "golang.org/x/image/draw"

	"easyPreparation_1.0/internal/path"
)

// GenerateConfig — 썸네일 생성 설정
type GenerateConfig struct {
	BackgroundPath string // 배경 이미지 경로

	// Legacy fields (backward compat)
	Title    string // DEPRECATED: DateLabel이 비어 있을 때 fallback
	SubTitle string // DEPRECATED: SermonTitle이 비어 있을 때 fallback

	// New layout fields
	DateLabel       string  // 상단 소 텍스트: "26.04.05 주일예배"
	SermonTitle     string  // 중앙 대 텍스트: 말씀 제목
	Scripture       string  // 하단 소 텍스트: 성경 참조
	LogoPath        string  // 선택: 로고 이미지 경로
	LogoPosition    string  // "top-left" | "top-right" | "bottom-left" | "bottom-right"
	LogoSizePercent float64 // 5~30 (캔버스 폭 %)

	FontName   string      // 폰트명 (빈 값 → NanumBrush 기본) — legacy
	TextStyles *TextStyles  // 영역별 텍스트 스타일 (nil이면 기본값)

	OutputPath string // 출력 경로
	Width      int    // 1280 (YouTube 표준)
	Height     int    // 720
}

// Generate — 배경 위에 한글 텍스트를 합성하여 썸네일 PNG 생성
func Generate(cfg GenerateConfig) (string, error) {
	if cfg.Width == 0 {
		cfg.Width = 1280
	}
	if cfg.Height == 0 {
		cfg.Height = 720
	}

	// 1. effective 값 결정 (backward compat)
	effectiveDateLabel := cfg.DateLabel
	if effectiveDateLabel == "" {
		effectiveDateLabel = cfg.Title
	}
	effectiveSermonTitle := cfg.SermonTitle
	if effectiveSermonTitle == "" {
		effectiveSermonTitle = cfg.SubTitle
	}
	effectiveScripture := cfg.Scripture

	// 2. 배경 이미지 로드
	bg, err := loadImage(cfg.BackgroundPath)
	if err != nil {
		// fallback: 어두운 그라데이션 배경
		solid := image.NewRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))
		draw.Draw(solid, solid.Bounds(), &image.Uniform{color.RGBA{25, 25, 55, 255}}, image.Point{}, draw.Src)
		bg = solid
	}

	// 3. 1280x720 cover-crop 스케일 (비율 유지, 중앙 크롭 — 세로 사진도 전체 채움)
	canvas := image.NewRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{color.RGBA{20, 20, 20, 255}}, image.Point{}, draw.Src)
	srcCrop := coverCropSrcBounds(bg.Bounds(), canvas.Bounds())
	xdraw.CatmullRom.Scale(canvas, canvas.Bounds(), bg, srcCrop, xdraw.Over, nil)

	// 4. 반투명 검정 오버레이 (rgba(0,0,0,0.3))
	overlay := image.NewUniform(color.RGBA{0, 0, 0, 76})
	draw.Draw(canvas, canvas.Bounds(), overlay, image.Point{}, draw.Over)

	// 5. TextStyles 결정
	ts := cfg.TextStyles
	if ts == nil {
		ts = DefaultTextStyles()
	}

	// 폰트 캐시 — 동일 폰트명은 한 번만 로드
	fontCache := map[string]*truetype.Font{}
	loadFont := func(name string) (*truetype.Font, error) {
		if f, ok := fontCache[name]; ok {
			return f, nil
		}
		f, err := loadFontByName(name)
		if err != nil {
			return nil, err
		}
		fontCache[name] = f
		return f, nil
	}

	headerFont, err := loadFont(ts.Header.FontName)
	if err != nil {
		return "", err
	}
	mainFont, err := loadFont(ts.Main.FontName)
	if err != nil {
		return "", err
	}
	footerFont, err := loadFont(ts.Footer.FontName)
	if err != nil {
		return "", err
	}

	headerColor := ParseHexColor(ts.Header.Color)
	mainColor := ParseHexColor(ts.Main.Color)
	footerColor := ParseHexColor(ts.Footer.Color)
	shadowColor := color.Color(color.RGBA{0, 0, 0, 180})

	// 6. 텍스트 렌더링 — 항상 주일예배 스타일 사용
	if effectiveDateLabel != "" {
		// 상단 소, Y = Height * 0.12
		dateY := int(float64(cfg.Height) * 0.12)
		drawTextCenteredReturnPos(canvas, headerFont, effectiveDateLabel, ts.Header.Size, cfg.Width, dateY+1, shadowColor)
		drawTextCenteredReturnPos(canvas, headerFont, effectiveDateLabel, ts.Header.Size, cfg.Width, dateY, headerColor)

		// 구분선
		faceHeader := truetype.NewFace(headerFont, &truetype.Options{Size: ts.Header.Size, DPI: 72})
		tw := measureString(faceHeader, effectiveDateLabel)
		faceHeader.Close()
		cx := cfg.Width / 2
		textLeft := cx - tw/2
		textRight := cx + tw/2
		gap := 20
		lineY := dateY - 25
		if textLeft-gap > 40 {
			drawHLine(canvas, 40, textLeft-gap, lineY, headerColor)
			drawHLine(canvas, 40, textLeft-gap, lineY+1, headerColor)
		}
		if textRight+gap < cfg.Width-40 {
			drawHLine(canvas, textRight+gap, cfg.Width-40, lineY, headerColor)
			drawHLine(canvas, textRight+gap, cfg.Width-40, lineY+1, headerColor)
		}
	}

	if effectiveSermonTitle != "" {
		// 중앙 대, Y = Height/2 + 20
		sermonY := cfg.Height/2 + 20
		drawTextCenteredWithShadow(canvas, mainFont, effectiveSermonTitle, ts.Main.Size, cfg.Width, sermonY, mainColor, shadowColor)
	} else if effectiveDateLabel != "" {
		// 말씀 제목이 없을 때: 날짜/예배명을 중앙에 크게
		titleY := cfg.Height/2 + 20
		drawTextCenteredWithShadow(canvas, mainFont, effectiveDateLabel, ts.Main.Size, cfg.Width, titleY, mainColor, shadowColor)
	}

	if effectiveScripture != "" {
		// 하단 소, Y = Height * 0.88
		scriptureY := int(float64(cfg.Height) * 0.88)
		drawTextCentered(canvas, footerFont, effectiveScripture, ts.Footer.Size, cfg.Width, scriptureY, footerColor, 0)
	}

	// 7. 로고 오버레이
	if cfg.LogoPath != "" {
		if err := composeLogo(canvas, cfg); err != nil {
			// 로고 실패는 무시하고 계속
			_ = err
		}
	}

	// 8. PNG 저장
	if err := os.MkdirAll(filepath.Dir(cfg.OutputPath), 0755); err != nil {
		return "", fmt.Errorf("출력 디렉토리 생성 실패: %w", err)
	}
	out, err := os.Create(cfg.OutputPath)
	if err != nil {
		return "", fmt.Errorf("출력 파일 생성 실패: %w", err)
	}
	defer out.Close()

	if err := png.Encode(out, canvas); err != nil {
		return "", fmt.Errorf("PNG 인코딩 실패: %w", err)
	}

	return cfg.OutputPath, nil
}

// composeLogo — 로고 이미지를 캔버스에 합성
func composeLogo(canvas *image.RGBA, cfg GenerateConfig) error {
	logoImg, err := loadImage(cfg.LogoPath)
	if err != nil {
		return fmt.Errorf("로고 로드 실패: %w", err)
	}

	// 로고 크기 결정
	sizePct := cfg.LogoSizePercent
	if sizePct <= 0 {
		sizePct = 18
	}
	logoW := int(float64(cfg.Width) * sizePct / 100)
	origBounds := logoImg.Bounds()
	origW := origBounds.Dx()
	origH := origBounds.Dy()
	if origW == 0 || origH == 0 {
		return fmt.Errorf("로고 크기가 0")
	}
	logoH := int(float64(logoW) * float64(origH) / float64(origW))

	// 리사이즈
	resized := image.NewRGBA(image.Rect(0, 0, logoW, logoH))
	xdraw.CatmullRom.Scale(resized, resized.Bounds(), logoImg, logoImg.Bounds(), xdraw.Over, nil)

	// 위치 결정
	const padding = 4
	var x, y int
	switch cfg.LogoPosition {
	case "top-left":
		x, y = padding, padding
	case "top-right":
		x, y = cfg.Width-logoW-padding, padding
	case "bottom-left":
		x, y = padding, cfg.Height-logoH-padding
	default: // "bottom-right"
		x, y = cfg.Width-logoW-padding, cfg.Height-logoH-padding
	}

	logoRect := image.Rect(x, y, x+logoW, y+logoH)
	draw.Draw(canvas, logoRect, resized, image.Point{}, draw.Over)
	return nil
}

// drawTextCenteredWithShadow — 텍스트 + 그림자 합성 (단방향, 테두리 없음)
func drawTextCenteredWithShadow(canvas *image.RGBA, f *truetype.Font, text string, size float64, canvasWidth, y int, col, shadowCol color.Color) {
	// 그림자 (오른쪽 하단 1방향 — 4방향 제거로 테두리 효과 없앰)
	drawTextCentered(canvas, f, text, size, canvasWidth, y+2, shadowCol, 2)
	// 본문
	drawTextCentered(canvas, f, text, size, canvasWidth, y, col, 0)
}

// loadFontByName — 이름으로 폰트 파일 로드
// 빈 값 또는 알 수 없는 폰트명은 NanumGothicBold(NanumGothic-800.ttf)로 fallback
// NanumBrush는 golang/freetype에서 한글 글리프를 잘못 렌더링하므로 선택지에서 제거됨
func loadFontByName(name string) (*truetype.Font, error) {
	fontFileMap := map[string]string{
		"NanumBrush":      "NanumBrush.ttf",
		"NanumGothic":     "NanumGothic-regular.ttf",
		"NanumGothicBold": "NanumGothic-800.ttf",
		"JacquesFrancois": "JacquesFrancois-regular.ttf",
	}
	fileName, ok := fontFileMap[name]
	if !ok {
		fileName = "NanumGothic-800.ttf"
	}

	// 1순위: execPath/public/font/ (프로덕션 — ExtractEmbeddedData가 복사한 파일)
	execPath := path.ExecutePath("easyPreparation")
	candidates := []string{
		filepath.Join(execPath, "public", "font", fileName),
		// 2순위: CWD 기준 ./public/font/ (개발 모드 — make dev 실행 시 CWD = 소스 루트)
		filepath.Join(".", "public", "font", fileName),
	}

	var fontData []byte
	var lastErr error
	for _, fontPath := range candidates {
		fontData, lastErr = os.ReadFile(fontPath)
		if lastErr == nil {
			break
		}
	}
	if lastErr != nil {
		return nil, fmt.Errorf("폰트 로드 실패 (%s): %w", fileName, lastErr)
	}
	f, err := truetype.Parse(fontData)
	if err != nil {
		return nil, fmt.Errorf("폰트 파싱 실패 (%s): %w", fileName, err)
	}
	return f, nil
}

// drawHLine — 수평선 그리기 (두께 1px, 두 번 호출해서 2px)
func drawHLine(canvas *image.RGBA, x1, x2, y int, col color.Color) {
	for x := x1; x <= x2; x++ {
		canvas.Set(x, y, col)
	}
}

// drawTextCenteredReturnPos — 텍스트를 중앙에 그리고 시작x와 텍스트 너비 반환
func drawTextCenteredReturnPos(canvas *image.RGBA, f *truetype.Font, text string, size float64, canvasWidth, y int, col color.Color) (startX, textWidth int) {
	face := truetype.NewFace(f, &truetype.Options{Size: size, DPI: 72})
	defer face.Close()
	textWidth = measureString(face, text)
	startX = (canvasWidth - textWidth) / 2

	ctx := freetype.NewContext()
	ctx.SetDPI(72)
	ctx.SetFont(f)
	ctx.SetFontSize(size)
	ctx.SetClip(canvas.Bounds())
	ctx.SetDst(canvas)
	ctx.SetSrc(image.NewUniform(col))
	ctx.SetHinting(font.HintingFull)
	pt := freetype.Pt(startX, y)
	ctx.DrawString(text, pt)
	return
}

func loadImage(p string) (image.Image, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(p))
	switch ext {
	case ".png":
		return png.Decode(f)
	case ".jpg", ".jpeg":
		return jpeg.Decode(f)
	default:
		img, _, err := image.Decode(f)
		return img, err
	}
}

func drawTextCentered(canvas *image.RGBA, f *truetype.Font, text string, size float64, canvasWidth, y int, col color.Color, shadowOffset int) {
	ctx := freetype.NewContext()
	ctx.SetDPI(72)
	ctx.SetFont(f)
	ctx.SetFontSize(size)
	ctx.SetClip(canvas.Bounds())
	ctx.SetDst(canvas)
	ctx.SetSrc(image.NewUniform(col))
	ctx.SetHinting(font.HintingFull)

	// 텍스트 너비 측정
	face := truetype.NewFace(f, &truetype.Options{Size: size, DPI: 72})
	defer face.Close()
	textWidth := measureString(face, text)

	x := (canvasWidth-textWidth)/2 + shadowOffset
	pt := freetype.Pt(x, y+shadowOffset)
	ctx.DrawString(text, pt)
}

// coverCropSrcBounds — CSS object-fit:cover 방식으로 소스 크롭 영역 계산
// 비율을 유지하면서 dst 전체를 채우는 src 크롭 영역 반환 (중앙 크롭)
func coverCropSrcBounds(srcBounds, dstBounds image.Rectangle) image.Rectangle {
	srcW := float64(srcBounds.Dx())
	srcH := float64(srcBounds.Dy())
	dstW := float64(dstBounds.Dx())
	dstH := float64(dstBounds.Dy())

	scaleX := dstW / srcW
	scaleY := dstH / srcH
	scale := math.Max(scaleX, scaleY) // cover: 더 큰 스케일 사용 → 전체 채움

	cropW := int(math.Round(dstW / scale))
	cropH := int(math.Round(dstH / scale))

	// 중앙 크롭
	offsetX := (srcBounds.Dx() - cropW) / 2
	offsetY := (srcBounds.Dy() - cropH) / 2

	return image.Rect(
		srcBounds.Min.X+offsetX,
		srcBounds.Min.Y+offsetY,
		srcBounds.Min.X+offsetX+cropW,
		srcBounds.Min.Y+offsetY+cropH,
	)
}

func measureString(face font.Face, s string) int {
	w := 0
	for _, r := range s {
		adv, ok := face.GlyphAdvance(r)
		if ok {
			w += int(math.Round(float64(adv) / 64.0))
		}
	}
	return w
}
