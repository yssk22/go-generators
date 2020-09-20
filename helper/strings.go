package helper

import (
	"unicode"
)

// ToSnakeCase converts the string to the one by snake case.
func ToSnakeCase(s string) string {
	if len(s) == 0 {
		return s
	}
	var runes = []rune(s)
	var str = []rune{unicode.ToLower(runes[0])}
	if len(runes) == 1 {
		return string(str)
	}
	for i := 1; i < len(runes)-1; i++ {
		previous := runes[i-1]
		current := runes[i]
		next := runes[i+1]
		if unicode.IsUpper(current) {
			if !unicode.IsUpper(previous) {
				str = append(str, '_')
			} else if unicode.IsLetter(next) && !unicode.IsUpper(next) {
				str = append(str, '_')
			}
			str = append(str, unicode.ToLower(current))
		} else {
			str = append(str, runes[i])
		}
	}
	str = append(str, unicode.ToLower(runes[len(runes)-1]))
	return string(str)
}
