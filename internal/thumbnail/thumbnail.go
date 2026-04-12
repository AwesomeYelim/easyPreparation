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
	Title          string // "4월 첫째주 주일예배"
	SubTitle       string // 선택: 말씀 제목 등
	OutputPath     string // 출력 경로
	Width          int    // 1280 (YouTube 표준)
	Height         int    // 720
}

// Generate — 배경 위에 한글 텍스트를 합성하여 썸네일 PNG 생성
func Generate(cfg GenerateConfig) (string, error) {
	if cfg.Width == 0 {
		cfg.Width = 1280
	}
	if cfg.Height == 0 {
		cfg.Height = 720
	}

	// 1. 배경 이미지 로드
	bg, err := loadImage(cfg.BackgroundPath)
	if err != nil {
		// fallback: 어두운 그라데이션 배경
		solid := image.NewRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))
		draw.Draw(solid, solid.Bounds(), &image.Uniform{color.RGBA{25, 25, 55, 255}}, image.Point{}, draw.Src)
		bg = solid
	}

	// 2. 1280x720 리사이즈
	canvas := image.NewRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))
	xdraw.CatmullRom.Scale(canvas, canvas.Bounds(), bg, bg.Bounds(), xdraw.Over, nil)

	// 3. 반투명 검정 오버레이 (rgba(0,0,0,0.3))
	overlay := image.NewUniform(color.RGBA{0, 0, 0, 76})
	draw.Draw(canvas, canvas.Bounds(), overlay, image.Point{}, draw.Over)

	// 4. 폰트 로드
	f, err := loadFont()
	if err != nil {
		return "", err
	}

	// 5. Title 텍스트 — 나눔손글씨 붓 (크게 + 두껍게, 그림자 없음)
	titleSize := 130.0
	titleY := cfg.Height/2 + 10
	white := color.Color(color.White)
	for _, off := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}, {0, 0}} {
		drawTextCentered(canvas, f, cfg.Title, titleSize, cfg.Width, titleY+off[1], white, off[0])
	}

	// 6. SubTitle (선택)
	if cfg.SubTitle != "" {
		subSize := 60.0
		subY := cfg.Height/2 + 100
		subWhite := color.Color(color.RGBA{255, 255, 255, 230})
		for _, off := range [][2]int{{-1, 0}, {1, 0}, {0, 0}} {
			drawTextCentered(canvas, f, cfg.SubTitle, subSize, cfg.Width, subY+off[1], subWhite, off[0])
		}
	}

	// 7. PNG 저장
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

func loadFont() (*truetype.Font, error) {
	execPath := path.ExecutePath("easyPreparation")
	fontPath := filepath.Join(execPath, "public", "font", "NanumBrush.ttf")
	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, fmt.Errorf("폰트 로드 실패: %w", err)
	}
	f, err := truetype.Parse(fontData)
	if err != nil {
		return nil, fmt.Errorf("폰트 파싱 실패: %w", err)
	}
	return f, nil
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
