package util

import (
	"strings"
	"unicode"
)

func ToUpperFirst(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

func ToSnakeCase(s string) string {
	for i := range s {
		if unicode.IsUpper(rune(s[i])) {
			s = strings.Join([]string{s[:i], ToLowerFirst(s[i:])}, "_")
		}
	}
	return s
}

func ToLowerFirst(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToLower(string(s[0])) + s[1:]
}

func IsInStringSlice(what string, where []string) bool {
	for _, item := range where {
		if item == what {
			return true
		}
	}
	return false
}

func FetchTags(strs []string, prefix string) (tags []string) {
	for _, comment := range strs {
		if strings.HasPrefix(comment, prefix) {
			tags = append(tags, strings.Split(strings.Replace(comment[len(prefix):], " ", "", -1), ",")...)
		}
	}
	return
}

func ToLower(str string) string {
	if len(str) > 0 && unicode.IsLower(rune(str[0])) {
		return str
	}
	for i := range str {
		if unicode.IsLower(rune(str[i])) {
			// Case, when only first char is upper.
			if i == 1 {
				return strings.ToLower(str[:1]) + str[1:]
			}
			return strings.ToLower(str[:i-1]) + str[i-1:]
		}
	}
	return strings.ToLower(str)
}

// Return last upper char in string or first char if no upper characters founded.
func LastUpperOrFirst(str string) string {
	for i := len(str) - 1; i >= 0; i-- {
		if unicode.IsUpper(rune(str[i])) {
			return string(str[i])
		}
	}
	return string(str[0])
}
