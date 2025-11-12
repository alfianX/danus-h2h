package function

import (
	"strings"
	"time"
)

func MaskPan(pan string) string {
	if pan != "" {
		length := len(pan)
		visibleCount := length / 4
		hiddenCount := length - (visibleCount * 2)

		mask := pan[:visibleCount] + strings.Repeat("*", hiddenCount) + pan[length-visibleCount:]

		return mask
	} else {
		return ""
	}
}

func PadRightZero(s string, totalLength int) string {
	if len(s) >= totalLength {
		return s
	}
	return s + strings.Repeat("0", totalLength-len(s))
}

func JulianDayNumber(t time.Time) int {
	year, month, day := t.Date()

	if month <= 2 {
		year--
		month += 12
	}

	a := year / 100
	b := 2 - a + a/4

	jdn := int(365.25*float64(year+4716)) +
		int(30.6001*float64(month+1)) +
		day + b - 1524

	return jdn
}
