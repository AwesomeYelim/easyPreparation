// Package sysopen — 시스템 기본 앱으로 URL/폴더를 여는 크로스 플랫폼 헬퍼.
//
// OS 분기가 파일마다 복붙되면 한 곳만 고쳐지는 사고가 반복되므로
// (예: Windows cmd /c start의 & 구분자 버그) 반드시 이 패키지를 사용할 것.
package sysopen

import (
	"os/exec"
	"runtime"
)

// URL — 시스템 기본 브라우저로 URL을 연다.
// Windows에서 cmd /c start는 URL의 &를 명령 구분자로 해석해 쿼리 파라미터가 잘리므로
// 셸을 거치지 않는 explorer.exe를 사용한다.
func URL(target string) error {
	return open(target)
}

// Folder — 파일 탐색기(Finder/Explorer)로 디렉터리를 연다.
func Folder(dir string) error {
	return open(dir)
}

func open(target string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", target)
	case "windows":
		cmd = exec.Command("explorer.exe", target)
	default:
		cmd = exec.Command("xdg-open", target)
	}
	return cmd.Start()
}
