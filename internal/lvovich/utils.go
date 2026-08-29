package lvovich

import "strings"

// endsWith — аналог JS String.prototype.endsWith.
func endsWith(str, search string) bool {
	return strings.HasSuffix(str, search)
}

// startsWith — аналог JS String.prototype.startsWith(str, search, pos).
func startsWith(str, search string, pos int) bool {
	if pos < 1 {
		pos = 0
	}
	return strings.HasPrefix(str[pos:], search)
}