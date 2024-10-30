package main

import (
	"fmt"
	"github.com/cascax/colorthief-go"
)

func main() {
	// 이미지에서 팔레트 추출 (6개의 색상)
	colors, err := colorthief.GetPaletteFromFile("./public/images/ppt_background.png", 6)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 추출한 색상 출력
	for i, color := range colors {
		r, g, b, _ := color.RGBA()
		fmt.Printf("Color %d: R:%d, G:%d, B:%d\n", i+1, r, g, b)
	}
}
