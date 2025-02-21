package format

import "fmt"

func IndexFormat(i, depth int) string {
	switch depth {
	case 1:
		return fmt.Sprintf("%d.", i+1) // 1., 2., 3.
	case 2:
		return fmt.Sprintf("%d)", i+1) // 1), 2), 3)
	case 3:
		return fmt.Sprintf("%c)", 'a'+i) // a), b), c)
	default:
		return fmt.Sprintf("%s)", intToRoman(i+1)) // i), ii), iii)
	}
}

// 숫자를 로마 숫자로 변환 (4단계 이상용)
func intToRoman(num int) string {
	vals := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	syms := []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}
	result := ""
	for i := 0; i < len(vals); i++ {
		for num >= vals[i] {
			num -= vals[i]
			result += syms[i]
		}
	}
	return result
}
