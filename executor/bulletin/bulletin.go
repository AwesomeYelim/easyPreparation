package main

import (
	"easyPreparation_1.0/executor/bulletin/contents"
	"easyPreparation_1.0/executor/bulletin/cover/colorPalette"
	"fmt"
)

func main() {
	colorList := colorPalette.GetColorWithSortByLuminance()
	highestLuminaceColor := colorList[len(colorList)-1]
	fmt.Println(highestLuminaceColor)
	contents.CreateContents()
}
