package quote

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func GetQuote(characters string) string {

	var dictB map[string]struct {
		Eng string `json:"eng"`
		Kor string `json:"kor"`
	}

	dict, err := os.ReadFile(filepath.Join("./config", "bible_dict.json"))
	err = json.Unmarshal(dict, &dictB)
	char := strings.Split(characters, " ")
	cover := ExpandRange(char[1])

	url := fmt.Sprintf("http://ibibles.net/quote.php?kor-%s/%s", dictB[char[0]].Eng, cover)

	// HTTP GET 요청 생성
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Failed to fetch data. Status code:", resp.StatusCode)
		return ""
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return ""
	}

	return StripTags(string(body))
}

func ExpandRange(input string) string {
	parts := strings.Split(input, ":")
	if len(parts) != 2 {
		return input
	}

	prefix := parts[0]
	rangePart := parts[1]

	rangeNumbers := strings.Split(rangePart, "-")
	if len(rangeNumbers) != 2 {
		return input
	}

	return fmt.Sprintf("%s:%s-%s:%s", prefix, rangeNumbers[0], prefix, rangeNumbers[1])
}

func StripTags(html string) string {
	html = strings.ReplaceAll(html, "\u003c", "<")
	html = strings.ReplaceAll(html, "\u003e", ">")
	html = strings.ReplaceAll(html, "&nbsp;", " ")
	html = strings.ReplaceAll(html, "\r\n", "\n")

	lines := strings.Split(html, "\n")
	var result []string

	for _, line := range lines {
		line = removeTags(line)

		if strings.HasPrefix(line, "Bible Quote") {
			continue
		}

		if strings.TrimSpace(line) != "" {
			result = append(result, strings.TrimSpace(line))
		}
	}

	return strings.Join(result, "\n")
}

func removeTags(input string) string {
	var output strings.Builder
	inTag := false
	for _, char := range input {
		if char == '<' {
			inTag = true
		} else if char == '>' {
			inTag = false
		} else if !inTag {
			output.WriteRune(char)
		}
	}
	return output.String()
}
