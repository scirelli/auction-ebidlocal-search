package notify

import (
	"context"
	"sync"
)

type NotificationMessageChanFilterer interface {
	Filter(context.Context, ...<-chan NotificationMessage) <-chan NotificationMessage
}

//NotificationMessageFilter returns true to allow the message to pass, false does not.
type NotificationMessageFilter func(NotificationMessage) bool

type NotificationMessageChanFilter func(context.Context, ...<-chan NotificationMessage) <-chan NotificationMessage

func (f NotificationMessageChanFilter) Filter(ctx context.Context, in ...<-chan NotificationMessage) <-chan NotificationMessage {
	return f(ctx, in...)
}

func NewFilter(filterFunc NotificationMessageFilter) NotificationMessageChanFilter {
	return NotificationMessageChanFilter(func(ctx context.Context, in ...<-chan NotificationMessage) <-chan NotificationMessage {
		var outChan = make(chan NotificationMessage)
		var wg sync.WaitGroup

		wg.Add(len(in))
		for _, c := range in {
			go func(c <-chan NotificationMessage) {
				defer wg.Done()
				for msg := range c {
					if filterFunc(msg) {
						select {
						case outChan <- msg:
						case <-ctx.Done():
						}
					}
				}
			}(c)
		}

		go func() {
			wg.Wait()
			close(outChan)
		}()

		return outChan
	})
}
