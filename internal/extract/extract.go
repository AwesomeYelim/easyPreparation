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
		Background: struct {
			Print        size.Box `json:"print"`
			Presentation size.Box `json:"presentation"`
		}{
			Print: size.Box{
				Width:  297.0,
				Height: 167.0,
			},
			Presentation: size.Box{
				Width:  350.0,
				Height: 210.0,
			},
		},
		InnerRectangle: size.Box{
			Width:  132,
			Height: 71,
		},
	},
	OutputPath: OutputPath{
		Bulletin: "output/bulletin",
		Lyrics:   "output/lyrics",
	},
}

func validateConfig(config *Config) {
	// Color 필드 체크
	if config.Color.BoxColor == "" {
		config.Color.BoxColor = defaultConfig.Color.BoxColor
	}
	if config.Color.LineColor == "" {
		config.Color.LineColor = defaultConfig.Color.LineColor
	}
	if config.Color.FontColor == "" {
		config.Color.FontColor = defaultConfig.Color.FontColor
	}
	if config.Color.DateColor == "" {
		config.Color.DateColor = defaultConfig.Color.DateColor
	}
	if config.Color.PrintColor == "" {
		config.Color.PrintColor = defaultConfig.Color.PrintColor
	}

	// Size 필드 체크
	if config.Size.Background.Print.Width == 0 {
		config.Size.Background.Print.Width = defaultConfig.Size.Background.Print.Width
	}
	if config.Size.Background.Print.Height == 0 {
		config.Size.Background.Print.Height = defaultConfig.Size.Background.Print.Height
	}
	if config.Size.Background.Presentation.Width == 0 {
		config.Size.Background.Presentation.Width = defaultConfig.Size.Background.Presentation.Width
	}
	if config.Size.Background.Presentation.Height == 0 {
		config.Size.Background.Presentation.Height = defaultConfig.Size.Background.Presentation.Height
	}

	if config.Size.InnerRectangle.Width == 0 {
		config.Size.InnerRectangle.Width = defaultConfig.Size.InnerRectangle.Width
	}
	if config.Size.InnerRectangle.Height == 0 {
		config.Size.InnerRectangle.Height = defaultConfig.Size.InnerRectangle.Height
	}

	// OutputPath 필드 체크
	if config.OutputPath.Bulletin == "" {
		config.OutputPath.Bulletin = defaultConfig.OutputPath.Bulletin
	}
	if config.OutputPath.Lyrics == "" {
		config.OutputPath.Lyrics = defaultConfig.OutputPath.Lyrics
	}
}

func ExtCustomOption(path string) (config Config) {
	custom, err := os.ReadFile(path)
	err = json.Unmarshal(custom, &config)

	if err != nil {
		log.Printf("%s Error :%s", path, err)
	}

	if reflect.DeepEqual(config, Config{}) {
		fmt.Println("config is empty")
	}
	// 유효성 검사 후 기본값 적용
	validateConfig(&config)

	return config
}
