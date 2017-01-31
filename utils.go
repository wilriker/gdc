package gdc

import (
	"fmt"
	"math"
	"strconv"
)

// Convert the given bytes count into a human readable form
func HumanReadableBytes(bytes int64) string {
	const unit = 1024
	const prefixes string = "KMGTPE"
	if bytes < unit {
		return strconv.FormatInt(bytes, 10) + "B"
	}
	b := float64(bytes)
	exp := math.Log(b / math.Log(unit))
	pre := prefixes[int64(exp)-1]
	return fmt.Sprintf("%.1f%c", b/math.Pow(unit, exp), pre)
}
