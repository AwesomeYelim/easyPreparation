package main

import (
	"easyPreparation_1.0/executor/bulletin/forPresentation"
	"easyPreparation_1.0/executor/bulletin/forPrint"
	"easyPreparation_1.0/executor/bulletin/scheme"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/gui"
	"easyPreparation_1.0/internal/path"
	"os"
	"path/filepath"
)

func main() {
	figmaInfo := gui.FigmaConnector()
	execPath, _ := os.Getwd()
	execPath = path.ExecutePath(execPath, "easyPreparation")
	configPath := filepath.Join(execPath, "config/custom.json")
	config := extract.ExtCustomOption(configPath)

	scheme.CreateScheme(figmaInfo)
	forPrint.CreatePrint(figmaInfo, execPath, config)
	forPresentation.CreatePresentation(figmaInfo, execPath, config)
}
