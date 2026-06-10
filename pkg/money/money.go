package money

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var yuanPattern = regexp.MustCompile(`^\d+(\.\d{1,2})?$`)

func YuanStringToCents(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if !yuanPattern.MatchString(value) {
		return 0, fmt.Errorf("invalid yuan amount %q", value)
	}

	parts := strings.SplitN(value, ".", 2)
	yuan, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid yuan amount %q: %w", value, err)
	}

	cents := uint64(0)
	if len(parts) == 2 {
		centPart := parts[1]
		if len(centPart) == 1 {
			centPart += "0"
		}
		cents, err = strconv.ParseUint(centPart, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid yuan amount %q: %w", value, err)
		}
	}

	maxCents := uint64(math.MaxInt64)
	if yuan > maxCents/100 || (yuan == maxCents/100 && cents > maxCents%100) {
		return 0, fmt.Errorf("yuan amount %q exceeds maximum cents value", value)
	}

	return int64(yuan*100 + cents), nil
}

func CentsToYuanString(cents int64) string {
	sign := ""
	amount := uint64(cents)
	if cents < 0 {
		sign = "-"
		amount = uint64(-(cents + 1)) + 1
	}
	return fmt.Sprintf("%s%d.%02d", sign, amount/100, amount%100)
}
