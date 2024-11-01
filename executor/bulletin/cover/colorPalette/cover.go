package colorPalette

import (
	"fmt"
	"github.com/cascax/colorthief-go"
	"image/color"
	"sort"
)

type ColorWithLuminance struct {
	Color     color.Color
	Luminance float64
}

type LsColorWithLuminance []ColorWithLuminance

func GetColorWithSortByLuminance() LsColorWithLuminance {
	// 색상과 명도를 저장할 슬라이스
	var colorLuminanceList LsColorWithLuminance
	colors := Init()
	colorLuminanceList.Run(colors)
	colorLuminanceList.Print()

	return colorLuminanceList
}

func Init() []color.Color {
	// 이미지에서 팔레트 추출 (4개의 색상)
	colors, err := colorthief.GetPaletteFromFile("./public/images/coverdesign.png", 4)
	if err != nil {
		fmt.Println("Error:", err)
		return []color.Color{}
	}
	return colors
}

func (lsc *LsColorWithLuminance) Run(colors []color.Color) {
	// 추출한 색상 출력 및 명도 계산
	for _, c := range colors {
		luminance := calculateLuminance(c)
		// 색상과 명도를 슬라이스에 추가
		*lsc = append(*lsc, ColorWithLuminance{Color: c, Luminance: luminance})
	}
	// 명도에 따라 색상 정렬
	sort.Sort(ByLuminance(*lsc))
}

func (lsc *LsColorWithLuminance) Print() {
	// 정렬된 색상 출력
	fmt.Println("\n정렬된 색상 (명도 순):")
	for i, cl := range *lsc {
		r, g, b, a := cl.Color.RGBA()
		ansiColor := fmt.Sprintf("\033[48;2;%d;%d;%d;%dm  \033[0m", r>>8, g>>8, b>>8, a>>8)
		fmt.Printf("Color %d: %s R:%d, G:%d, B:%d, A:%d, Luminance: %.2f\n", i+1, ansiColor, r>>8, g>>8, b>>8, a>>8, cl.Luminance)
	}
}
