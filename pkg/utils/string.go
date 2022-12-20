package utils

import "strings"

func Lines(src string) []string {
	src = strings.TrimSpace(src)
	return strings.Split(src, "\n")
}

func SplitKeyVal(src string, sep string) (string, string) {
	r := strings.SplitN(src, sep, 2)
	key := strings.TrimSpace(r[0])
	val := ""
	if len(r) > 1 {
		val = strings.TrimSpace(r[1])
	}
	return key, val
}
