package scheme

import (
	"easyPreparation_1.0/internal/figma/get"
)

func CreateScheme(figmaInfo *get.Info) {
	figmaInfo.AssembledNodes = figmaInfo.GetFrames("forEdit")
	figmaInfo.GetResource("main_worship")
}
