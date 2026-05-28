package attacks

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

var suffixes = []string{
	"1", "2", "3", "12", "123", "1234", "12345",
	"0", "00", "007", "69", "99", "2023", "2024", "2025",
	"!", "@", "#", "$", "1!", "123!", "!1",
}

var prefixes = []string{"1", "123", "0"}

func ApplyRules(word string) []string {
	if utf8.RuneCountInString(word) > 30 {
		return []string{word}
	}
	seen := make(map[string]struct{}, 100)
	add := func(s string) {
		if s != "" {
			seen[s] = struct{}{}
		}
	}

	lower := strings.ToLower(word)
	cap1 := capitalize(lower)
	l1 := leet1(lower)
	l2 := leet2(lower)

	add(word); add(lower); add(strings.ToUpper(word)); add(cap1)
	add(l1); add(l2); add(reverseStr(lower)); add(lower+lower)

	for _, base := range []string{lower, cap1, l1, l2} {
		for _, suf := range suffixes {
			add(base + suf)
		}
	}
	for _, pre := range prefixes {
		add(pre + lower); add(pre + cap1)
	}

	out := make([]string, 0, len(seen))
	for s := range seen { out = append(out, s) }
	return out
}

func capitalize(s string) string {
	if s == "" { return s }
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

func reverseStr(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 { r[i], r[j] = r[j], r[i] }
	return string(r)
}

func leet1(s string) string {
	var b strings.Builder
	for _, c := range s {
		switch c {
		case 'a': b.WriteByte('@')
		case 'e': b.WriteByte('3')
		case 'i': b.WriteByte('!')
		case 'o': b.WriteByte('0')
		case 's': b.WriteByte('$')
		case 't': b.WriteByte('7')
		default:  b.WriteRune(c)
		}
	}
	return b.String()
}

func leet2(s string) string {
	var b strings.Builder
	for _, c := range s {
		switch c {
		case 'a': b.WriteByte('4')
		case 'e': b.WriteByte('3')
		case 'i': b.WriteByte('1')
		case 'o': b.WriteByte('0')
		case 's': b.WriteByte('5')
		case 't': b.WriteByte('+')
		default:  b.WriteRune(c)
		}
	}
	return b.String()
}
