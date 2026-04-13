package forPresentation

import (
	"easyPreparation_1.0/internal/bulletin/define"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/presentation"
	"easyPreparation_1.0/internal/types"
	"easyPreparation_1.0/internal/utils"
	"encoding/json"
	"fmt"
	"github.com/jung-kurt/gofpdf/v2"
	"os"
	"path/filepath"
	"strings"
)

type PdfInfo struct {
	*define.PdfInfo
}

func (pi PdfInfo) Create() {
	config := extract.ConfigMem
	outputDir := filepath.Join(pi.ExecPath, "data", "templates", "display")
	_ = utils.CheckDirIs(outputDir)

	// 배경 이미지 로드 (없어도 계속 진행 — 흰 배경으로 대체)
	pathInfo := make(map[string]string)
	if imgFiles, err := os.ReadDir(outputDir); err == nil {
		for _, img := range imgFiles {
			if img.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(img.Name()))
			if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
				pathInfo[strings.TrimSuffix(img.Name(), ext)] = filepath.Join(outputDir, img.Name())
			}
		}
	}

	instanceSize := gofpdf.SizeType{
		Wd: config.Classification.Bulletin.Presentation.Width,
		Ht: config.Classification.Bulletin.Presentation.Height,
	}
	objPdf := presentation.New(instanceSize)
	objPdf.Config = config.Classification.Bulletin.Presentation
	objPdf.PdfInfo = pi.PdfInfo

	var contents []types.WorshipInfo

	worshipContents, err := os.ReadFile(filepath.Join(pi.ExecPath, "config", pi.Target+".json"))
	if err != nil {
		fmt.Printf("[forPresentation] config 읽기 실패: %v\n", err)
		return
	}
	if err = json.Unmarshal(worshipContents, &contents); err != nil {
		fmt.Printf("[forPresentation] config 파싱 실패: %v\n", err)
		return
	}

	for _, con := range contents {
		objPdf.Title = con.Title

		if strings.Contains(objPdf.Title, "성시교독") {
			continue
		}

		objPdf.AddPage()

		if path, ok := pathInfo[objPdf.Title]; ok {
			objPdf.Path = filepath.Join(outputDir, filepath.Base(path))
			objPdf.CheckImgPlaced(objPdf.Path, 0)
		}

		objPdf.MarkName()

		if strings.Contains(con.Info, "edit") || strings.Contains(con.Info, "notice") {
			objPdf.ForEdit(con, config)
		}
	}

	outputBtPath := filepath.Join(pi.ExecPath, config.OutputPath.Bulletin, "presentation")
	_ = utils.CheckDirIs(outputBtPath)
	bulletinPath := filepath.Join(outputBtPath, pi.OutputFilename)
	if err = objPdf.OutputFileAndClose(bulletinPath); err != nil {
		fmt.Printf("[forPresentation] PDF 저장 실패: %v\n", err)
	}
}
