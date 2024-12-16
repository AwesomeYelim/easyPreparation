package forPresentation

import (
	"easyPreparation_1.0/internal/colorPalette"
	"easyPreparation_1.0/internal/date"
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/figma/get"
	"easyPreparation_1.0/internal/presentation"
	"easyPreparation_1.0/pkg"
	"encoding/json"
	"fmt"
	"github.com/jung-kurt/gofpdf/v2"
	"os"
	"path/filepath"
	"strings"
)

func CreatePresentation(figmaInfo *get.Info, execPath string, config extract.Config) {
	outputDir := filepath.Join(execPath, config.OutputPath.Bulletin, "presentation", "tmp")
	_ = pkg.CheckDirIs(outputDir)

	defer func() {
		_ = os.RemoveAll(outputDir)
	}()

	figmaInfo.GetFigmaImage(outputDir, "forShowing")
	highestLuminaceColor := colorPalette.HexToRGBA(config.Color.BoxColor)

	bulletinSize := gofpdf.SizeType{
		Wd: config.Size.Background.Width,
		Ht: config.Size.Background.Height,
	}
	yearMonth, weekFormatted := date.SetDateTitle()

	objPdf := presentation.New(bulletinSize)

	outputFilename := fmt.Sprintf("%s_%s.pdf", yearMonth, weekFormatted)

	var contents []get.Children
	custom, err := os.ReadFile(filepath.Join(execPath, "config", "main_worship.json"))
	err = json.Unmarshal(custom, &contents)

	for _, con := range contents {
		title := strings.Split(con.Title, "_")[1]

		if path, ok := figmaInfo.PathInfo[title]; ok {
			imgPath := filepath.Join(outputDir, filepath.Base(path))
			objPdf.AddPage()
			objPdf.CheckImgPlaced(bulletinSize, imgPath, 0)

			if strings.Contains(con.Info, "edit") {
				if strings.HasPrefix(con.Title, "2_") {
					objPdf.SetText(20, highestLuminaceColor)
					var subContents []string
					// 공백 제거
					trimmedText := pkg.RemoveEmptyLines(con.Obj)
					lines := strings.Split(trimmedText, "\n")
					for i := 0; i < len(lines); i++ {
						if strings.HasPrefix(lines[i], "Bible Quote") {
							lines[i] = strings.TrimPrefix(lines[i], "Bible Quote")
						}
						lines[i] = removeLineNumberPattern(lines[i])
						subContents = append(subContents, lines[i])
					}
					for i := 0; i < len(subContents); i += 4 {
						if i > 0 {
							objPdf.AddPage()
							objPdf.CheckImgPlaced(bulletinSize, imgPath, 0)
						}

						if i+4 < len(lines) {
							subContents = append(subContents, lines[i]+"\n"+lines[i+1]+"\n"+lines[i+2]+"\n"+lines[i+3])
						} else {
							subContents = append(subContents, lines[i])
						}

						var textW float64 = 250
						var textH float64 = 20
						// 텍스트 추가 (내용 설정)
						objPdf.SetXY((bulletinSize.Wd-textW)/2, (bulletinSize.Ht-textH)/2)
						objPdf.MultiCell(textW, textH, subContents[i], "", "C", false)
					}

				} else {
					objPdf.SetText(27, highestLuminaceColor)
					objPdf.WriteText(148.5, 110, con.Content)
				}
			}
		}
	}

	outputBtPath := filepath.Join(execPath, config.OutputPath.Bulletin, "presentation")
	_ = pkg.CheckDirIs(outputBtPath)
	bulletinPath := filepath.Join(outputBtPath, outputFilename)
	err = objPdf.OutputFileAndClose(bulletinPath)
	if err != nil {
		fmt.Printf("PDF 저장 중 에러 발생: %v", err)
	}

}
func removeLineNumberPattern(text string) string {
	// 앞에서부터 숫자와 콜론(:)이 나오면 이를 제거
	i := 0
	for i < len(text) && (text[i] >= '0' && text[i] <= '9' || text[i] == ':') {
		i++
	}
	return strings.TrimSpace(text[i:])
}
