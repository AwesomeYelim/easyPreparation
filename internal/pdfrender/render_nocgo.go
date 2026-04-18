//go:build !cgo

package pdfrender

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// PDFToImages converts a PDF file to numbered PNG images (1.png, 2.png, ...) in outDir.
// Falls back to Ghostscript when CGO is disabled (server mode).
func PDFToImages(pdfPath, outDir string, dpi int) error {
	pngPattern := filepath.Join(outDir, "%d.png")
	dpiFlag := fmt.Sprintf("-r%d", dpi)

	gsPath := resolveGhostscript()
	cmd := exec.Command(gsPath, "-sDEVICE=pngalpha", "-o", pngPattern, dpiFlag, pdfPath)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gs 변환 실패: %s — %w", string(output), err)
	}
	return nil
}

func resolveGhostscript() string {
	switch runtime.GOOS {
	case "windows":
		for _, name := range []string{"gswin64c", "gswin32c", "gs"} {
			if _, err := exec.LookPath(name); err == nil {
				return name
			}
		}
		// 일반적인 Windows 설치 경로
		for _, p := range []string{
			`C:\Program Files\gs\gs10.04.0\bin\gswin64c.exe`,
			`C:\Program Files\gs\gs10.03.1\bin\gswin64c.exe`,
			`C:\Program Files (x86)\gs\gs10.04.0\bin\gswin32c.exe`,
		} {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
		return "gswin64c"
	default:
		// macOS: Homebrew 경로 우선, 없으면 PATH에서 탐색
		for _, p := range []string{
			"/opt/homebrew/bin/gs",
			"/usr/local/bin/gs",
		} {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
		if path, err := exec.LookPath("gs"); err == nil {
			return path
		}
		return "gs"
	}
}
