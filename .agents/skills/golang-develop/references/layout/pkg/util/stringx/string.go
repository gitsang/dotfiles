package stringx

import "strings"

func TrimSplit(s string, sep string) []string {
	parts := strings.Split(s, sep)
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			items = append(items, item)
		}
	}
	return items
}
