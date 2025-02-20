package extract

import (
	"easyPreparation_1.0/internal/classification"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
)

type OutputPath struct {
	Bulletin string `json:"bulletin"`
	Lyrics   string `json:"lyrics"`
}

type Config struct {
	Classification classification.BackgroundInfo `json:"classification"`
	OutputPath     OutputPath                    `json:"outputPath"`
}

var defaultConfig = Config{
	Classification: classification.BackgroundInfo{
		Bulletin: classification.Bulletin{
			Print: classification.ResultInfo{
				Size: classification.Size{
					Width:  1409.0,
					Height: 996.0,
				},
				FontInfo: classification.FontInfo{
					FontSize:   50.0,
					FontOption: "Nanum Gothic",
				},
				InnerRectangle: classification.Size{
					Width:  510,
					Height: 860,
				},
				Color: classification.Color{
					BoxColor:   "#F8F3EA",
					LineColor:  "#BEA07C",
					FontColor:  "#BEA07C",
					DateColor:  "#FFFFFF",
					PrintColor: "#8B7F71",
				},
			},
			Presentation: classification.ResultInfo{
				Size: classification.Size{
					Width:  1409.0,
					Height: 996.0,
				},
				FontInfo: classification.FontInfo{
					FontSize:   100.0,
					FontOption: "Nanum Gothic",
				},
				InnerRectangle: classification.Size{
					Width:  1278,
					Height: 640,
				},
				Color: classification.Color{
					BoxColor:   "#F8F3EA",
					LineColor:  "#BEA07C",
					FontColor:  "#BEA07C",
					DateColor:  "#FFFFFF",
					PrintColor: "#8B7F71",
				},
			},
		},
		Lyrics: classification.Lyrics{Presentation: classification.ResultInfo{
			Size: classification.Size{
				Width:  1409.0,
				Height: 792.0,
			},
			FontInfo: classification.FontInfo{
				FontSize:   130.0,
				FontOption: "Nanum Gothic",
			},
			InnerRectangle: classification.Size{
				Width:  1278,
				Height: 640,
			},
			Color: classification.Color{
				BoxColor:   "#F8F3EA",
				LineColor:  "#BEA07C",
				FontColor:  "#BEA07C",
				DateColor:  "#FFFFFF",
				PrintColor: "#8B7F71",
			},
		}},
	},
	OutputPath: OutputPath{
		Bulletin: "output/bulletin",
		Lyrics:   "output/lyrics",
	},
}

func fillDefaults(dst, def reflect.Value) {
	for i := 0; i < dst.NumField(); i++ {
		field := dst.Field(i)
		defField := def.Field(i)

		// 필드를 설정할 수 없는 경우 건너뜀
		if !field.CanSet() {
			continue
		}

		//  또 다른 struct라면 재귀적으로 처리
		if field.Kind() == reflect.Struct {
			fillDefaults(field, defField)
		} else {
			if field.IsZero() {
				field.Set(defField)
			}
		}
	}
}

func validateConfig(config *Config) {
	fillDefaults(reflect.ValueOf(config).Elem(), reflect.ValueOf(defaultConfig))
}

func scaleFloats(v reflect.Value) {
	switch v.Kind() {
	case reflect.Float64:
		if v.CanSet() {
			v.SetFloat(v.Float() / 2)
		}
	case reflect.Ptr:
		if !v.IsNil() {
			scaleFloats(v.Elem())
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			scaleFloats(v.Field(i))
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			scaleFloats(v.Index(i))
		}
	default:

	}
}

func ExtCustomOption(path string) {
	custom, err := os.ReadFile(path)
	err = json.Unmarshal(custom, &ConfigMem)

	if err != nil {
		log.Printf("%s Error :%s", path, err)
	}

	if reflect.DeepEqual(ConfigMem, Config{}) {
		fmt.Println("ConfigMem is empty")
	}
	// 유효성 검사 후 기본값 적용
	validateConfig(&ConfigMem)
	//scaleFloats(reflect.ValueOf(&ConfigMem).Elem()) // 1/2 크기로 사용

}

var (
	ConfigMem Config
)
