package util

import (
	"fmt"
	"strings"
)

func HumanDuration(sec int64) string {
	h := sec / 3600
	m := (sec % 3600) / 60
	s := sec % 60
	parts := []string{}
	if h > 0 {
		parts = append(parts, fmt.Sprintf("%d小时", h))
	}
	if m > 0 || (h > 0 && s > 0) {
		parts = append(parts, fmt.Sprintf("%d分", m))
	}
	if s > 0 || (h == 0 && m == 0) {
		parts = append(parts, fmt.Sprintf("%d秒", s))
	}
	return strings.Join(parts, "")
}