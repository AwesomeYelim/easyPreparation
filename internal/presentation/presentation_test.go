package presentation

import (
	"easyPreparation_1.0/internal/extract"
	"easyPreparation_1.0/internal/path"
	"easyPreparation_1.0/pkg"
	"fmt"
	"github.com/jung-kurt/gofpdf/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetBody(t *testing.T) {
	execPath, _ := os.Getwd()
	execPath = path.ExecutePath(execPath, "easyPreparation")
	configPath := filepath.Join(execPath, "config/custom.json")
	config := extract.ExtCustomOption(configPath)

	bulletinSize := gofpdf.SizeType{
		Wd: config.Size.Background.Presentation.Width,
		Ht: config.Size.Background.Presentation.Height,
	}
	var textSize float64 = 25
	var textW float64 = 230

	objPdf := New(bulletinSize)
	objPdf.FullSize = bulletinSize
	objPdf.Path = filepath.Join(execPath, "public", "images", "ppt_background.png")
	objPdf.Contents = strings.Split("84:1 (고라 자손의 시. 영장으로 깃딧에 맞춘 노래) 만군의 여호와여, 주의 장막이 어찌 그리 사랑스러운지요\n84:2 내 영혼이 여호와의 궁정을 사모하여 쇠약함이여 내 마음과 육체가 생존하시는 하나님께 부르짖나이다\n84:3 나의 왕, 나의 하나님, 만군의 여호와여, 주의 제단에서 참새도 제 집을 얻고 제비도 새끼 둘 보금자리를 얻었나이다\n84:4 주의 집에 거하는 자가 복이 있나이다 저희가 항상 주를 찬송하리이다 셀라\n84:5 주께 힘을 얻고 그 마음에 시온의 대로가 있는 자는 복이 있나이다\n84:6 저희는 눈물 골짜기로 통행할 때에 그 곳으로 많은 샘의 곳이 되게 하며 이른 비도 은택을 입히나이다\n84:7 저희는 힘을 얻고 더 얻어 나아가 시온에서 하나님 앞에 각기 나타나리이다\n84:8 만군의 하나님 여호와여, 내 기도를 들으소서 야곱의 하나님이여, 귀를 기울이소서 (셀라)\n84:9 우리 방패이신 하나님이여, 주의 기름 부으신 자의 얼굴을 살펴보옵소서\n84:10 주의 궁정에서 한 날이 다른 곳에서 천 날보다 나은즉 악인의 장막에 거함보다 내 하나님 문지기로 있는 것이 좋사오니\n84:11 여호와 하나님은 해요 방패시라 여호와께서 은혜와 영화를 주시며 정직히 행하는 자에게 좋은 것을 아끼지 아니하실 것임이니이다\n84:12 만군의 여호와여, 주께 의지하는 자는 복이 있나이다", "\n")
	objPdf.setBody(textW, textSize, 3)

	outputBtPath := filepath.Join(execPath, config.OutputPath.Bulletin, "presentation")
	_ = pkg.CheckDirIs(outputBtPath)
	bulletinPath := filepath.Join(outputBtPath, "test.pdf")
	err := objPdf.OutputFileAndClose(bulletinPath)
	if err != nil {
		fmt.Printf("PDF 저장 중 에러 발생: %v", err)
	}

}
