package sorted

import (
	"os"
	"sort"
	"strconv"
	"strings"
)

func ToIntSort[T []os.DirEntry](files T, trimPrefix string, trimSuffix string) {
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

		aN := extractNumber(files[a].Name())
		bN := extractNumber(files[b].Name())

		return aN < bN
	})
}
