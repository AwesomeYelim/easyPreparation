package build

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func UiBuild(uiBuildPath, destPath string) (buildFolder string) {

	// 환경 변수 확인 -> dev 모드에서만 UI 빌드 실행
	env := os.Getenv("APP_ENV")
	if env == "dev" {
		err := runPnpmBuild(uiBuildPath)
		if err != nil {
			fmt.Printf("Error running pnpm build: %v\n", err)
			os.Exit(1)
		}

		buildFolder = filepath.Join(uiBuildPath, "build")
		if err = copyDirectory(buildFolder, destPath); err != nil {
			log.Fatalf("Failed to copy build folder: %v", err)
		}

		// build 폴더 내의 index.html 경로 설정
		htmlFilePath := filepath.Join(buildFolder, "index.html")

		// 파일이 존재하는지 확인
		if _, err := os.Stat(htmlFilePath); os.IsNotExist(err) {
			log.Fatalf("Failed to find the HTML file at: %v", htmlFilePath)
		}
	} else {
		fmt.Println("Skipping UI build (not in dev mode).")
		return
	}
	// 빌드된 React 프로젝트의 경로

	return buildFolder
}

func copyDirectory(srcDir, destDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 대상 경로 설정
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(destDir, relPath)

		// 디렉토리이면 생성
		if d.IsDir() {
			return os.MkdirAll(targetPath, os.ModePerm)
		}

		// 파일이면 복사
		return copyFile(path, targetPath)
	})
}

func copyFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}

// runPnpmBuild 실행 함수
func runPnpmBuild(projectPath string) error {
	fmt.Println("Building UI with pnpm...")

	// pnpm build 명령어 실행
	cmd := exec.Command("pnpm", "build")
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run pnpm build: %w", err)
	}

	fmt.Println("UI build completed successfully.")
	return nil
}
