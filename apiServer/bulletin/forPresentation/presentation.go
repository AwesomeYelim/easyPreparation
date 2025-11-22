package forPresentation

import (
	"easyPreparation_1.0/apiServer/bulletin/define"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/figma/get"
	"easyPreparation_1.0/internal/gui"
	"easyPreparation_1.0/internal/presentation"
	"easyPreparation_1.0/pkg"
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
	outputDir := filepath.Join(pi.ExecPath, config.OutputPath.Bulletin, "presentation", "tmp")
	//_ = pkg.CheckDirIs(outputDir)
	//
	//defer func() {
	//	_ = os.RemoveAll(outputDir)
	//}()

	//pi.FigmaInfo.GetFigmaImage(outputDir, "forShowing")

	pi.FigmaInfo = &get.Info{
		PathInfo: make(map[string]string),
	}

	imgFiles, _ := os.ReadDir(outputDir)
	for _, img := range imgFiles {
		pi.FigmaInfo.PathInfo[strings.TrimSuffix(img.Name(), ".png")] = filepath.Join(outputDir, img.Name())
	}

	instanceSize := gofpdf.SizeType{
		Wd: config.Classification.Bulletin.Presentation.Width,
		Ht: config.Classification.Bulletin.Presentation.Height,
	}
	objPdf := presentation.New(instanceSize)
	objPdf.Config = config.Classification.Bulletin.Presentation
	objPdf.PdfInfo = pi.PdfInfo

	var contents []gui.WorshipInfo

	worshipContents, err := os.ReadFile(filepath.Join(pi.ExecPath, "config", pi.Target+".json"))
	err = json.Unmarshal(worshipContents, &contents)

	for _, con := range contents {
		objPdf.Title = con.Title

		if path, ok := pi.FigmaInfo.PathInfo[objPdf.Title]; ok {
			objPdf.Path = filepath.Join(outputDir, filepath.Base(path))
			if !strings.Contains(objPdf.Title, "성시교독") {
				objPdf.AddPage()
				objPdf.CheckImgPlaced(objPdf.Path, 0)
				objPdf.MarkName()
			}
			if strings.Contains(con.Info, "edit") || strings.Contains(con.Info, "notice") {
				objPdf.ForEdit(con, config)
			}
		}
	}

	outputBtPath := filepath.Join(pi.ExecPath, config.OutputPath.Bulletin, "presentation")
	_ = pkg.CheckDirIs(outputBtPath)
	bulletinPath := filepath.Join(outputBtPath, pi.OutputFilename)
	err = objPdf.OutputFileAndClose(bulletinPath)
	if err != nil {
		fmt.Printf("PDF 저장 중 에러 발생: %v", err)
	}

}
