package utils

import (
	"fmt"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	p = message.NewPrinter(language.English)
)

func FormatNumber(num float64, digits int) string {
	format := fmt.Sprintf("%%.%df", digits)
	return p.Sprintf(format, num)
}

func Signum(num float64) int {
	switch {
	case num < 0:
		return -1
	case num > 0:
		return 1
	default:
		return 0
	}
}
