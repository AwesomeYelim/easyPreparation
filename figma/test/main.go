package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func updateNodeImage(figmaAPIKey, fileKey, nodeID, imageURL string) error {
	// 요청 데이터 구성
	requestBody := map[string]interface{}{
		"requests": []map[string]interface{}{
			{
				"op":   "UPDATE",
				"path": fmt.Sprintf("/document/children/0/children/%s/fills", nodeID),
				"value": []map[string]interface{}{
					{
						"type":     "IMAGE",
						"imageRef": imageURL, // 이미지 URL
					},
				},
			},
		},
	}

	// JSON 직렬화
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	// API URL
	url := fmt.Sprintf("https://api.figma.com/v1/files/%s", fileKey)

	// HTTP 요청 생성
	client := &http.Client{}
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// 요청 헤더 설정
	req.Header.Add("X-Figma-Token", figmaAPIKey)
	req.Header.Add("Content-Type", "application/json")

	// 요청 전송
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update node, status: %d", resp.StatusCode)
	}

	log.Println("Node image updated successfully!")
	return nil
}

func main() {
	// API 키, 파일 키, 노드 ID, 이미지 URL
	figmaAPIKey := ""
	fileKey := ""
	nodeID := "2:2" // Rectangle 1의 ID
	imageURL := "https://your-image-url.com/image.png"

	// 이미지 변경 요청
	err := updateNodeImage(figmaAPIKey, fileKey, nodeID, imageURL)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
