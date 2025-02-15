package extract

import (
	"easyPreparation_1.0/internal/size"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
)

type Color struct {
	BoxColor   string `json:"boxColor"`
	LineColor  string `json:"lineColor"`
	FontColor  string `json:"fontColor"`
	DateColor  string `json:"dateColor"`
	PrintColor string `json:"printColor"`
}
type OutputPath struct {
	Bulletin string `json:"bulletin"`
	Lyrics   string `json:"lyrics"`
}

type Config struct {
	Color      Color      `json:"color"`
	Size       size.Size  `json:"size"`
	OutputPath OutputPath `json:"outputPath"`
}

var defaultConfig = Config{
	Color: Color{
		BoxColor:   "#F8F3EA",
		LineColor:  "#BEA07C",
		FontColor:  "#BEA07C",
		DateColor:  "#FFFFFF",
		PrintColor: "#8B7F71",
	},
	Size: size.Size{
		Background: size.Background{
			Print: size.ResultInfo{
				Width:    1409.0,
				Height:   996.9,
				FontSize: 40.0,
			},
			Presentation: size.ResultInfo{
				Width:    1409.0,
				Height:   880.0,
				FontSize: 100.0,
			},
		},
		InnerRectangle: size.ResultInfo{
			Width:  584,
			Height: 279,
		},
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

}

var (
	ConfigMem Config
)
