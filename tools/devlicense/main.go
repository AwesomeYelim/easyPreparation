// devlicense — 개발/테스트용 Pro 라이선스를 data/license.json에 기록한다.
//
// 사용:
//
//	go run ./tools/devlicense/           # Pro, 무기한
//	go run ./tools/devlicense/ free      # Free 플랜으로 초기화
//	go run ./tools/devlicense/ enterprise # Enterprise 플랜
//
// 또는: make dev-license
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"easyPreparation_1.0/internal/license"
)

func main() {
	// 플랜 인수 파싱 (기본: pro)
	plan := license.PlanPro
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "free":
			plan = license.PlanFree
		case "enterprise":
			plan = license.PlanEnterprise
		case "pro":
			plan = license.PlanPro
		default:
			fmt.Fprintf(os.Stderr, "알 수 없는 플랜: %s (사용 가능: free / pro / enterprise)\n", os.Args[1])
			os.Exit(1)
		}
	}

	// 키 + 서명 생성 (만료일 없음 = 무기한)
	key, sig := license.GenerateTestKey(plan, time.Time{})

	now := time.Now()
	info := license.LicenseInfo{
		LicenseKey:   key,
		Plan:         plan,
		DeviceID:     "",  // 오프라인 캐시라 DeviceID 없어도 동작
		ChurchID:     1,
		IssuedAt:     now,
		ExpiresAt:    time.Time{}, // 무기한
		LastVerified: now,
		Signature:    sig,
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSON 직렬화 실패: %v\n", err)
		os.Exit(1)
	}

	outPath := filepath.Join("data", "license.json")
	if err := os.MkdirAll("data", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "data/ 디렉토리 생성 실패: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(outPath, data, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "파일 쓰기 실패: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 개발용 %s 라이선스 생성 완료\n", strings.ToUpper(string(plan)))
	fmt.Printf("   파일: %s\n", outPath)
	fmt.Printf("   키:   %s\n", key)
	fmt.Printf("   앱을 재시작하면 적용됩니다.\n")
}
