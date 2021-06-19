package funcUtils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGlobalThrottleFunc(t *testing.T) {
	var DefaultThrottle = ThrottleFuncFactory(10)
	t.Run("Should call the function for all values", func(t *testing.T) {
		var sum, max int = 0, 5

		var callMe = func(i int) {
			sum += i
			time.Sleep(time.Duration(i) * 500 * time.Millisecond)
		}

		for i := 1; i <= max; i++ {
			DefaultThrottle(func(v ...interface{}) {
				callMe(v[0].(int))
			}, i)
		}
		time.Sleep(200 * time.Millisecond)
		assert.Equal(t, 15, sum)
	})

	t.Run("Should call the function 10 times before throttling.", func(t *testing.T) {
		var sum, max int = 0, 15

		var callMe = func(i int) {
			sum += i
		}

		for i := 1; i <= max; i++ {
			DefaultThrottle(func(v ...interface{}) {
				callMe(v[0].(int))
			}, i)
		}
		time.Sleep(200 * time.Millisecond)
		assert.Equal(t, 120, sum)
	})

	t.Run("Should call the function with no params", func(t *testing.T) {
		var sum, max int = 0, 15

		var callMe = func(i int) {
			sum += i
		}

		for i := 1; i <= max; i++ {
			DefaultThrottle(func(v ...interface{}) {
				callMe(i)
			})
		}
		time.Sleep(200 * time.Millisecond)
	})
}
