package function

import "strings"

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
