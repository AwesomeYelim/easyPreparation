package googleCloud

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

func GetGoogleCloudInfo(folder, target, outputPath string) {
	srv := Login()
	// 상위 폴더 ID와 하위 폴더 이름 설정
	parentFolderID := "16gLC6lcmxfQRTbHVlhSQwqYhR-kijuEm" // 교회 이름 폴더 ID

	// 1. 하위 폴더 ID 검색
	subFolderID := getSubFolderID(srv, parentFolderID, folder)
	if subFolderID == "" {
		log.Fatalf("하위 폴더 '%s'를 찾을 수 없습니다.", folder)
	}

	// 2. 하위 폴더에서 파일 검색
	file := getFileInFolder(srv, subFolderID, target)
	if file == nil {
		log.Fatalf("파일 '%s'을(를) 하위 폴더 '%s'에서 찾을 수 없습니다.", target, folder)
	}

	// 3. 파일 다운로드
	outputPath = fmt.Sprintf(filepath.Join(outputPath, "%s"), file.Name)
	downloadFile(srv, file.Id, outputPath)
}

// 0. Google Drive API 클라이언트 초기화
func Login() (srv *drive.Service) {
	ctx := context.Background()
	path, err := filepath.Abs("./config/auth.json")
	srv, err = drive.NewService(ctx, option.WithCredentialsFile(path))
	if err != nil {
		log.Fatalf("Google Drive 클라이언트 초기화 실패: %v", err)
	}
	return srv
}

func getSubFolderID(srv *drive.Service, parentFolderID, subFolderName string) string {
	// 현재 폴더에서 하위 폴더를 검색
	query := fmt.Sprintf("'%s' in parents and mimeType = 'application/vnd.google-apps.folder' and trashed = false", parentFolderID)
	foldersListCall := srv.Files.List().Q(query).Fields("files(id, name)")
	folderList, _ := foldersListCall.Do()

	// 현재 폴더의 하위 폴더 중 이름이 일치하는 폴더를 찾음
	for _, folder := range folderList.Files {
		if folder.Name == subFolderName {
			return folder.Id
		}
	}

	// 일치하는 폴더가 없으면 하위 폴더를 재귀적으로 탐색
	for _, folder := range folderList.Files {
		foundID := getSubFolderID(srv, folder.Id, subFolderName)
		if foundID != "" {
			return foundID
		}
	}

	// 폴더를 찾을 수 없는 경우 빈 문자열 반환
	return ""
}

func getFileInFolder(srv *drive.Service, folderID, fileName string) *drive.File {
	query := fmt.Sprintf("'%s' in parents and name = '%s' and trashed = false", folderID, fileName)
	filesListCall := srv.Files.List().Q(query).Fields("files(id, name)")
	fileList, err := filesListCall.Do()
	if err != nil {
		log.Fatalf("파일 검색 실패: %v", err)
	}

	if len(fileList.Files) == 0 {
		return nil
	}

	// 첫 번째 파일 반환
	return fileList.Files[0]
}

func downloadFile(srv *drive.Service, fileID, outputPath string) {
	// Google Drive에서 파일 콘텐츠 다운로드
	res, err := srv.Files.Get(fileID).Download()
	if err != nil {
		log.Fatalf("파일 다운로드 실패: %v", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	// 로컬 파일로 저장
	outFile, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("로컬 파일 생성 실패: %v", err)
	}

	defer func() {
		_ = outFile.Close()
	}()

	_, err = io.Copy(outFile, res.Body)
	if err != nil {
		log.Fatalf("파일 저장 실패: %v", err)
	}

	fmt.Printf("파일 다운로드 성공: %s\n", outputPath)
}
