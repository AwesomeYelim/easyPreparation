package main

import (
	"fmt"
	"gocv.io/x/gocv"
	"math"
)

func main() {
	img := gocv.IMRead("test.jpg", gocv.IMReadGrayScale)
	if img.Empty() {
		fmt.Println("Image load failed!")
		return
	}
	defer img.Close()

	// 이진화
	gocv.Threshold(img, &img, 200, 255, gocv.ThresholdBinary)

	// 엣지 검출
	edges := gocv.NewMat()
	defer edges.Close()
	gocv.Canny(img, &edges, 50, 150)

	// 선 검출
	lines := gocv.NewMat()
	defer lines.Close()

	_ = gocv.HoughLinesP(edges, &lines, 1, math.Pi/180, 50)

	fmt.Printf("Total Lines detected: %d\n", lines.Rows())

	// 선 정보 꺼내기
	for i := 0; i < lines.Rows(); i++ {
		line := lines.GetVeciAt(i, 0) // (Veci 타입)
		if isHorizontal(line) {
			fmt.Printf("Horizontal Line: %+v\n", line)
		}
	}
}

func isHorizontal(line gocv.Veci) bool {
	dx := line[2] - line[0]
	dy := line[3] - line[1]
	if dy == 0 {
		return true
	}
	slope := float64(dy) / float64(dx)
	return slope > -0.1 && slope < 0.1
}
