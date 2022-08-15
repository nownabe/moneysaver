package main

import (
	"fmt"
	"strings"
)

func humanize(n int64) string {
	s := fmt.Sprint(n)
	l := (len(s) + 3 - 1) / 3
	parts := make([]string, l)

	for i := 0; i < l; i++ {
		start := len(s) - (l-i)*3
		end := start + 3
		if start < 0 {
			start = 0
		}
		parts[i] = s[start:end]
	}
	return "¥" + strings.Join(parts, ",")
}
