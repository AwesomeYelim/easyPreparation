//go:build !cgo

package pdfrender

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// PDFToImages converts a PDF file to numbered PNG images (1.png, 2.png, ...) in outDir.
// Falls back to Ghostscript when CGO is disabled (server mode).
func PDFToImages(pdfPath, outDir string, dpi int) error {
	pngPattern := fmt.Sprintf("%s/%s", outDir, "%d.png")
	dpiFlag := fmt.Sprintf("-r%d", dpi)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		gsPath := "gswin64c"
		if _, err := exec.LookPath("gswin64c"); err != nil {
			if _, err2 := exec.LookPath("gswin32c"); err2 == nil {
				gsPath = "gswin32c"
			} else {
				gsPath = "gs"
			}
		}
		cmd = exec.Command(gsPath, "-sDEVICE=pngalpha", "-o", pngPattern, dpiFlag, pdfPath)
	default:
		gsPath := "/opt/homebrew/bin/gs"
		if _, err := os.Stat(gsPath); err != nil {
			gsPath = "gs"
		}
		cmd = exec.Command(gsPath, "-sDEVICE=pngalpha", "-o", pngPattern, dpiFlag, pdfPath)
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gs 변환 실패: %s — %w", string(output), err)
	}
	return nil
}
