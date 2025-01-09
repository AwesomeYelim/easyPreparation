package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func main() {
	// Google Drive API 클라이언트 초기화
	ctx := context.Background()
	path, _ := filepath.Abs("./config/auth.json")
	srv, err := drive.NewService(ctx, option.WithCredentialsFile(path))
	if err != nil {
		log.Fatalf("Google Drive 클라이언트 초기화 실패: %v", err)
	}

	// 검색하려는 폴더 ID와 파일 이름 설정
	folderID := "19NSzi4n2WAZSZZ1nzidEJ26G84ZUgz6z"
	fileName := "001.pdf"

	// 폴더 내에서 파일 검색
	query := fmt.Sprintf("'%s' in parents and name = '%s'", folderID, fileName) // trashed 조건 제거
	filesListCall := srv.Files.List().Q(query).Fields("files(id, name, trashed)")
	fileList, err := filesListCall.Do()
	if err != nil {
		log.Fatalf("파일 검색 실패: %v", err)
	}

	if len(fileList.Files) == 0 {
		log.Fatalf("'%s' 파일이 폴더에서 발견되지 않았습니다.", fileName)
	}

	// 파일 ID 가져오기
	file := fileList.Files[0] // 첫 번째 검색 결과 사용
	fmt.Printf("다운로드할 파일: %s (ID: %s)\n", file.Name, file.Id)

	// 파일 다운로드
	outputPath := fmt.Sprintf("./%s", file.Name)
	downloadFile(srv, file.Id, outputPath)
}

func downloadFile(srv *drive.Service, fileID, outputPath string) {
	// Google Drive에서 파일 콘텐츠 다운로드
	res, err := srv.Files.Get(fileID).Download()
	if err != nil {
		log.Fatalf("파일 다운로드 실패: %v", err)
	}
	defer res.Body.Close()

	// 로컬 파일로 저장
	outFile, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("로컬 파일 생성 실패: %v", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, res.Body)
	if err != nil {
		log.Fatalf("파일 저장 실패: %v", err)
	}

	fmt.Printf("파일 다운로드 성공: %s\n", outputPath)
}
