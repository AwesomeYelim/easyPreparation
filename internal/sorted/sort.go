package sorted

import (
	"os"
	"sort"
	"strconv"
	"strings"
)

func ToIntSort[T []os.DirEntry | []string](files T, trimPrefix string, trimSuffix string) {
	sort.Slice(files, func(a, b int) bool {
		extractNumber := func(name string) int {
			parts := strings.TrimPrefix(name, trimPrefix)
			parts = strings.TrimSuffix(name, trimSuffix)
			num, err := strconv.Atoi(parts)
			if err == nil {
				return num
			}
			return 0
		}

		var aN, bN int

		switch fileTyped := any(files).(type) {
		case []os.DirEntry:
			aN = extractNumber(fileTyped[a].Name())
			bN = extractNumber(fileTyped[b].Name())
			break
		case []string:
			aN = extractNumber(fileTyped[a])
			bN = extractNumber(fileTyped[b])
		}

		return aN < bN
	})
}
