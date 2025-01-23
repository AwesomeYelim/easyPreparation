package extract

import (
	"easyPreparation_1.0/internal/size"
	"encoding/json"
	"log"
	"os"
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

func ExtCustomOption(path string) (config Config) {
	custom, err := os.ReadFile(path)
	err = json.Unmarshal(custom, &config)

	if err != nil {
		log.Printf("%s Error :%s", path, err)
		config.Color.BoxColor = "#FFFFFF"
	}

	return config
}
