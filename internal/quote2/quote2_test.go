package quote2

import (
	"fmt"
	"testing"
)

// 📌 **네트워크 테스트: GetQuote 실행 후 크롤링 결과 출력**
func TestGetQuote(t *testing.T) {
	// 📌 실제 사용할 forUrl 예제
	testCases := []string{
		"1/1:1-1:3", // 창세기 1장 1~3절\
	}

	for _, forUrl := range testCases {
		t.Run(forUrl, func(t *testing.T) {
			fmt.Println("\n===== 크롤링 요청 =====")
			fmt.Printf("요청 구절: %s\n", forUrl)

			// 📌 크롤링 실행
			GetQuote(forUrl)

			fmt.Println("========================")
		})
	}
}
