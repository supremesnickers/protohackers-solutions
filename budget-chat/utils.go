package main

import (
	"strings"
)

func stripNewline(str string) string {
	return strings.ReplaceAll(str, "\n", "")
}
