package main

import (
	"regexp"
	"strconv"
	"strings"
)

type Canonicalizer struct {
	replacer            *strings.Replacer
	keepDotInFloatRegex *regexp.Regexp
}

func NewCanonicalizer() *Canonicalizer {
	removeChars := strings.Split(".,:;[]{}#?@\\~`\"'", "")
	// we use string to upper to prevent using operators
	//	removeOperators := strings.Split("$env,not,or,and,matches,contains,startsWith,endsWith", ",")

	replaceOldNew := make([]string, 0, len(removeChars))

	for _, char := range removeChars {
		replaceOldNew = append(replaceOldNew, char, "_r$"+strconv.Itoa(int(char[0]))+"$r_")
	}

	return &Canonicalizer{
		replacer: strings.NewReplacer(replaceOldNew...),
		// https://regex101.com/r/O1Eabf/2
		keepDotInFloatRegex: regexp.MustCompile(`((?:[^\w_$]|^)\d+)_r\$46\$r_([\dEe]+(?:[^\w_$]|$))`),
	}
}

func (c *Canonicalizer) Canonicalize(s string) string {
	return c.keepDotInFloatRegex.ReplaceAllString(
		c.replacer.Replace(strings.ToUpper(s)),
		"$1.$2",
	)
}
