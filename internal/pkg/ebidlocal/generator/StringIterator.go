package ebidlocal

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
type StringIterator interface {
	Next() (string, bool)
}

type StringGenerator interface {
	Generator() StringIterator
}

type StringIteratorFunc func() (string, bool)

func (s StringIteratorFunc) Next() (string, bool) {
	return s()
}

type SliceStringGenerator []string

func (ssg SliceStringGenerator) Generator() StringIterator {
	var i int
	return StringIteratorFunc(func() (string, bool) {
		if i >= len(ssg) {
			return "", false
		}
		v := ssg[i]
		i++
		return v, true
	})
}
