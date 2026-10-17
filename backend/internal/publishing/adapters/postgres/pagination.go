package postgres

import "strconv"

func offsetValue(values []int) int {
	if len(values) == 0 {
		return 0
	}
	return max(0, min(values[0], 1000000))
}
func pageOffset(values []int) string { return strconv.Itoa(offsetValue(values)) }
