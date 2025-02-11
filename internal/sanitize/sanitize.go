package sanitize

import "strings"

func FileName(fileName string) string {
	replacer := strings.NewReplacer(
		"<", "_",
		">", "_",
		":", "_",
		`"`, "_",
		"/", "_",
		`\`, "_",
		"|", "_",
		"?", "_",
		"*", "_",
	)

	safeName := replacer.Replace(fileName)
	return strings.TrimSpace(safeName)
}
