// Package pgdsn — tools/ 스크립트용 PostgreSQL DSN 로더.
// 자격증명을 소스에 두지 않고 환경변수 또는 gitignore된 config/db.json에서 읽는다.
package pgdsn

import (
	"encoding/json"
	"fmt"
	"os"
)

// Load — PG_DSN 환경변수 → config/db.json의 "dsn" 키 순서로 DSN을 찾는다.
func Load() (string, error) {
	if dsn := os.Getenv("PG_DSN"); dsn != "" {
		return dsn, nil
	}
	const configPath = "config/db.json"
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("PG_DSN 환경변수도 없고 %s 도 읽을 수 없음: %w", configPath, err)
	}
	var cfg struct {
		DSN string `json:"dsn"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("%s 파싱 오류: %w", configPath, err)
	}
	if cfg.DSN == "" {
		return "", fmt.Errorf("%s 에 \"dsn\" 키가 비어 있음", configPath)
	}
	return cfg.DSN, nil
}

// MustLoad — Load 실패 시 메시지를 출력하고 종료한다.
func MustLoad() string {
	dsn, err := Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "PostgreSQL DSN 로드 실패:", err)
		os.Exit(1)
	}
	return dsn
}
