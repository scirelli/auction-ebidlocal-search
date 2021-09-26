package stringutils

import (
	"regexp"
	"strings"
)

var matchPunctuation = regexp.MustCompile(`[[:punct:]]`)

func FilterEmpty(s []string) (out []string) {
	out = make([]string, 0, len(s))
	for _, item := range s {
		if item != "" {
			out = append(out, item)
		}
	}
	return
}

func SliceToDict(s []string) (dict map[string]struct{}) {
	dict = make(map[string]struct{})
	for _, w := range s {
		dict[w] = struct{}{}
	}
	return
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
