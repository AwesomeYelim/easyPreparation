package main

import (
	"easyPreparation_1.0/executor/bulletin/forPresentation"
	"easyPreparation_1.0/executor/bulletin/forPrint"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/gui"
	"easyPreparation_1.0/internal/path"
	"os"
	"path/filepath"
)

func main() {
	execPath, _ := os.Getwd()
	execPath = path.ExecutePath(execPath, "easyPreparation")
	target, figmaInfo := gui.SetBulletinGui(execPath)
	configPath := filepath.Join(execPath, "config/custom.json")
	extract.ExtCustomOption(configPath)

	forPrint.CreatePrint(figmaInfo, target, execPath)
	forPresentation.CreatePresentation(figmaInfo, target, execPath)
}
