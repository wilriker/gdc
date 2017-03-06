package gdc

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// HumanReadableBytes converts the given bytes count into a human readable form
func HumanReadableBytes(bytes uint64) string {
	const unit = 1024
	const prefixes string = "KMGTPE"
	if bytes < unit {
		return strconv.FormatUint(bytes, 10) + "B"
	}
	b := float64(bytes)
	exp := math.Log(b / math.Log(unit))
	pre := prefixes[int64(exp)-1]
	return fmt.Sprintf("%.1f%c", b/math.Pow(unit, exp), pre)
}

// Exists returns whether the given file or directory exists or not
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// FixPath repairs a path if necessary by preprending slash and removing trailing slash
func FixPath(path string) string {
	// TODO This can probably be improved
	p := strings.TrimSpace(path)
	if len(p) > 0 {
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		if len(p) > 1 && strings.HasSuffix(p, "/") {
			p = p[:len(p)-1]
		}
	}
	return p
}
