package utils

import "strings"

func AddHex(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	if strings.HasPrefix(s, "0x") {
		return s
	}
	return strings.ToLower("0x" + s)
}
