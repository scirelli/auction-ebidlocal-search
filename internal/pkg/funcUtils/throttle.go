package funcUtils

//Throttler throttler interface. Throttlers are used to throttle multiple executions of a function.
type Throttler interface {
	Throttle(func(...interface{}), ...interface{})
}

type ThrottleFunc func(func(...interface{}), ...interface{})

func (tf ThrottleFunc) Throttle(fnc func(...interface{}), params ...interface{}) {
	tf(fnc, params...)
}

//ThrottleFuncFactory create throttle functions that only allow max concurrent executions.
func ThrottleFuncFactory(max int) ThrottleFunc {
	var semephoreChan = make(chan struct{}, max)
	return func(fnc func(...interface{}), params ...interface{}) {
		semephoreChan <- struct{}{}
		go func() {
			defer func() { <-semephoreChan }()
			fnc(params...)
		}()
	}
}

//MaxThrottle max throttles are set per instance.
type MaxThrottle struct {
	semephoreChan chan struct{}
}

func (mt *MaxThrottle) Throttle(fnc func(...interface{}), params ...interface{}) {
	mt.semephoreChan <- struct{}{}
	go func() {
		defer func() { <-mt.semephoreChan }()
		fnc(params...)
	}()
}

//NewMaxThrottle create an instance of MaxThrottle with the maximum number of conncurrent calls allowed, already set.
func NewMaxThrottle(maxConcurrentCalls int) *MaxThrottle {
	return &MaxThrottle{
		semephoreChan: make(chan struct{}, maxConcurrentCalls),
	}
}
