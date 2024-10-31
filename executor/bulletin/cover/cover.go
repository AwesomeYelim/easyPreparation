package main

import (
	"fmt"
	"github.com/cascax/colorthief-go"
)

func main() {
	// 이미지에서 팔레트 추출 (6개의 색상)
	colors, err := colorthief.GetPaletteFromFile("./public/images/cover.png", 4)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 추출한 색상 출력
	for i, color := range colors {
		r, g, b, a := color.RGBA()
		ansiColor := fmt.Sprintf("\033[48;2;%d;%d;%d;%dm  \033[0m", r>>8, g>>8, b>>8, a>>8)
		fmt.Printf("Color %d: %s R:%d, G:%d, B:%d A:%d \n", i+1, ansiColor, r>>8, g>>8, b>>8, a>>8)
	}
}
