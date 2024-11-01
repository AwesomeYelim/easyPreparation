package colorPalette

import "image/color"

type ByLuminance []ColorWithLuminance

func (a ByLuminance) Len() int           { return len(a) }
func (a ByLuminance) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByLuminance) Less(i, j int) bool { return a[i].Luminance < a[j].Luminance }

func calculateLuminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	return 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
}
