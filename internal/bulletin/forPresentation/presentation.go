package forPresentation

import (
	"easyPreparation_1.0/internal/bulletin/define"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/handlers"
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

	// 배경 이미지 로드
	pathInfo := make(map[string]string)
	loadPathInfo := func(dir string) {
		if imgFiles, err := os.ReadDir(dir); err == nil {
			for _, img := range imgFiles {
				if img.IsDir() {
					continue
				}
				ext := strings.ToLower(filepath.Ext(img.Name()))
				if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
					key := strings.TrimSuffix(img.Name(), ext)
					if _, exists := pathInfo[key]; !exists {
						pathInfo[key] = filepath.Join(dir, img.Name())
					}
				}
			}
		}
	}
	loadPathInfo(outputDir)

	// templates에 없는 키는 data/defaults/bulletin/presentation/ 에서 보충 (복사 안 함)
	// → display.go는 data/templates/display/만 스캔하므로 defaults 복사 시 display bgImage에 영향을 줌
	defaultsDir := filepath.Join(pi.ExecPath, "data", "defaults", "bulletin", "presentation")
	loadPathInfo(defaultsDir)

	instanceSize := gofpdf.SizeType{
		Wd: config.Classification.Bulletin.Presentation.Width,
		Ht: config.Classification.Bulletin.Presentation.Height,
	}
	objPdf := presentation.New(instanceSize)
	objPdf.Config = config.Classification.Bulletin.Presentation
	objPdf.PdfInfo = pi.PdfInfo
	objPdf.SetAutoPageBreak(false, 0) // presentation은 슬라이드 단위 — MultiCell 자동 페이지 추가 방지

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

		hasBackground := false
		if _, ok := pathInfo[con.Title]; ok {
			hasBackground = true
		}
		hasContent := strings.Contains(con.Info, "edit") || strings.Contains(con.Info, "notice")

		// 배경도 없고 내용도 없는 항목은 슬라이드 생략 (흰화면 방지)
		if !hasBackground && !hasContent {
			continue
		}

		handlers.BroadcastProgress("Presentation", 1, fmt.Sprintf("[슬라이드] %s", con.Title))

		// 성시교독: ForEdit 내부에서 PNG 페이지를 자체 추가 — 외부 AddPage/MarkName 불필요 (흰화면 방지)
		if con.Title == "성시교독" {
			objPdf.ForEdit(con, config)
		} else {
			objPdf.AddPage()
			if hasBackground {
				objPdf.Path = pathInfo[objPdf.Title]
				objPdf.CheckImgPlaced(objPdf.Path, 0)
			}
			objPdf.MarkName()
			objPdf.DrawSlideHeader(con.Lead, con.Title)
			if hasContent {
				objPdf.ForEdit(con, config)
			}
		}

		// 축도 이후 항목 제외
		if con.Title == "축도" {
			break
		}
	}

	outputBtPath := filepath.Join(pi.ExecPath, config.OutputPath.Bulletin, "presentation")
	_ = utils.CheckDirIs(outputBtPath)
	bulletinPath := filepath.Join(outputBtPath, pi.OutputFilename)
	if err = objPdf.OutputFileAndClose(bulletinPath); err != nil {
		fmt.Printf("[forPresentation] PDF 저장 실패: %v\n", err)
	}
}
