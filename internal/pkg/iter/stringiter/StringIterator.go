package stringiter

// type StringIterator interface {
// 	Next() bool
// 	Value() string
// 	Key() int
// }
// type StringGeneratorer interface {
// 	Generator() StringIterator
// }

// type stringIteratorData struct {
// 	index int
// 	slice []string
// }

// func (sg *stringIteratorData) Next() bool {
// 	sg.index++
// 	return sg.index < len(sg.slice)
// }
// func (sg *stringIteratorData) Value() string {
// 	return sg.slice[sg.index]
// }
// func (sg *stringIteratorData) Key() int {
// 	return sg.index
// }

// type StringGenerator []string

// func (sg StringGenerator) Generator() StringIterator {
// 	return &stringIteratorData{
// 		index: -1,
// 		slice: sg,
// 	}
// }

//----------------------------------------------------------------
type Iterator interface {
	Next() (string, bool)
}

type Iterable interface {
	Iterator() Iterator
}

type IteratorFunc func() (string, bool)

func (s IteratorFunc) Next() (string, bool) {
	return s()
}

type SliceStringIterator []string

func (ssg SliceStringIterator) Iterator() Iterator {
	var i int
	return IteratorFunc(func() (string, bool) {
		if i >= len(ssg) {
			return "", false
		}
		v := ssg[i]
		i++
		return v, true
	})
}

type ChanStringIterator chan string

func (csg ChanStringIterator) Iterator() Iterator {
	return IteratorFunc(func() (string, bool) {
		value, ok := <-csg
		return value, ok
	})
}
