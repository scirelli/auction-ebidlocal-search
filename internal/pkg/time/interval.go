package time

import (
	"context"
	gotime "time"
)

//DoEvery execute a function every time interval.
func DoEvery(ctx context.Context, d gotime.Duration, f func(gotime.Time)) error {
	ticker := gotime.NewTicker(d)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case x := <-ticker.C:
			f(x)
		}
	}
}

func SetTimeout(f func(), t gotime.Duration) (done chan struct{}) {
	done = make(chan struct{})
	ticker := gotime.NewTicker(t)

	go func() {
		defer close(done)
		defer ticker.Stop()

		select {
		case <-ticker.C:
			f()
			return
		case <-done:
			break
		}

	}()

	return done
}
