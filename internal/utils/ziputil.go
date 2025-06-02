package ziputil

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
)

func CreateZipBufferFromFiles(filePaths []string, fileNames []string) ([]byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	for i, path := range filePaths {
		err := addFileToZipWriter(zipWriter, path, fileNames, i)
		if err != nil {
			return nil, err
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func addFileToZipWriter(zipWriter *zip.Writer, filePath string, fileNames []string, index int) error {
	file, info, err := openAndStatFile(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	header, err := createZipHeader(info, filePath, fileNames, index)
	if err != nil {
		return err
	}

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	return copyFileToZip(writer, file)
}

func openAndStatFile(path string) (*os.File, os.FileInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, err
	}

	return file, info, nil
}

func createZipHeader(info os.FileInfo, filePath string, fileNames []string, index int) (*zip.FileHeader, error) {
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return nil, err
	}

	if fileNames != nil && index < len(fileNames) && fileNames[index] != "" {
		header.Name = fileNames[index]
	} else {
		header.Name = filepath.Base(filePath)
	}

	header.Method = zip.Deflate
	return header, nil
}

func copyFileToZip(writer io.Writer, file *os.File) error {
	_, err := io.Copy(writer, file)
	return err
}
