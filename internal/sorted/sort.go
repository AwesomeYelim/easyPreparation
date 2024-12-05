package sorted

import (
	"os"
	"sort"
	"strconv"
	"strings"
)

func ToIntSort[T []os.DirEntry | []string](files T, trimPrefix string, trimSuffix string, flag int) {
	sort.Slice(files, func(a, b int) bool {
		extractNumber := func(name string) float64 {
			parts := strings.TrimPrefix(name, trimPrefix)
			if flag == 1 {
				if strings.Contains(name, trimSuffix) {
					parts = strings.Split(name, trimSuffix)[0]
				}
			} else {
				parts = strings.TrimSuffix(name, trimSuffix)
			}
			num, err := strconv.ParseFloat(parts, 64)
			if err == nil {
				return num
			}
			return 0
		}

		var aN, bN float64

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
