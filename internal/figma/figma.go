package figma

import (
	"easyPreparation_1.0/internal/figma/get"
	"fmt"
)

func New(token *string, key *string, execPath string) (*get.Info, error) {
	if token == nil || *token == "" {
		return nil, fmt.Errorf("Figma 토큰이 없습니다")
	}
	if key == nil || *key == "" {
		return nil, fmt.Errorf("Figma 파일 키가 없습니다")
	}
	return &get.Info{
		Token:    token,
		Key:      key,
		ExecPath: execPath,
		PathInfo: make(map[string]string),
	}, nil
}
