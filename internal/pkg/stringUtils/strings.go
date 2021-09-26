package stringutils

import (
	"regexp"
	"strings"
)

var matchPunctuation = regexp.MustCompile(`[[:punct:]]`)

func FilterEmpty(s []string) []string {
	return s
}

func SliceToDict(s []string) map[string]struct{} {
	dict := make(map[string]struct{})
	for _, w := range s {
		dict[w] = struct{}{}
	}
	return dict
}

func ToLower(s []string) []string {
	o := make([]string, len(s))
	for i, w := range s {
		o[i] = strings.ToLower(w)
	}
	return o
}

func StripPunctuation(s []string) []string {
	o := make([]string, len(s))
	for i, w := range s {
		o[i] = string(matchPunctuation.ReplaceAll([]byte(w), []byte("")))
	}
	return o
}
