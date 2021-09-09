package stringiter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type SliceStringIterTestCase struct {
	Expected []string
}

func TestSliceStringIterator(t *testing.T) {
	var tests map[string]SliceStringIterTestCase = map[string]SliceStringIterTestCase{
		"Should loop over a slice of strings": SliceStringIterTestCase{
			Expected: []string{"a", "b", "c"},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			iter := SliceStringIterator(test.Expected).Iterator()
			for _, expected := range test.Expected {
				result, ok := iter.Next()
				assert.Equalf(t, expected, result, "'%v' not equal '%v'", result, expected)
				assert.Equalf(t, true, ok, "'%v' not equal '%v'", ok, true)
			}
			result, ok := iter.Next()
			assert.Equalf(t, false, ok, "'%v' not equal '%v'", ok, false)
			assert.Equalf(t, "", result, "'%v' not equal '%v'", result, "")
		})
	}
}
