package generator

import (
	"sync"
)

//https://ewencp.org/blog/golang-iterators/index.html
//IncrementingNumber returns a function that each time called it will return an incrementing number.
func IncrementingNumber(start, step int) func() int {
	var i int = start
	var mu sync.Mutex

	if step == 0 {
		step = 1
	}

	return func() int {
		var cur int = i
		mu.Lock()
		defer mu.Unlock()
		i += step
		return cur
	}
}

type AutoInc struct {
	i    int
	step int
	mu   sync.Mutex
}

func NewAutoInc(start, step int) *AutoInc {
	if step == 0 {
		step = 1
	}

	return &AutoInc{
		i:    start,
		step: step,
	}
}

func (i *AutoInc) NewID() int {
	i.mu.Lock()
	defer i.mu.Unlock()
	defer i.inc()

	return i.i
}

func (i *AutoInc) inc() {
	i.i += i.step
}
