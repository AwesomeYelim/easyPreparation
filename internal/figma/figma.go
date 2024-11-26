package figma

import (
	"easyPreparation_1.0/internal/figma/get"
	"flag"
	"github.com/torie/figma"
	"os"
)

func New(token *string, key *string, execPath string) (node *get.Info) {
	if *token == "" || *key == "" {
		flag.Usage()
		os.Exit(-1)
	}
	return &get.Info{
		Client:   figma.New(*token),
		Token:    token,
		Key:      key,
		ExecPath: execPath,
	}
}
