package notify

import (
	"sync"
)

type Deduper interface {
	Enqueue(msgChan <-chan NotificationMessage)
}

type DedupeQueue struct {
	mux sync.RWMutex
	set map[string]struct{}
}

func NewDedupeQueue() *DedupeQueue {
	return &DedupeQueue{
		set: make(map[string]struct{}),
	}
}

//DedupeChan fan in
func (q *DedupeQueue) Enqueue(msgChan <-chan NotificationMessage) <-chan NotificationMessage {
	var empty struct{}
	var dedupedChan = make(chan NotificationMessage)
	var wg sync.WaitGroup

	go func() {
		for msg := range msgChan {
			q.mux.RLock()
			_, ok := q.set[msg.User.ID+msg.WatchlistID]
			q.mux.RUnlock()
			if !ok {
				q.mux.Lock()
				q.set[msg.User.ID+msg.WatchlistID] = empty
				q.mux.Unlock()
				wg.Add(1)
				go func(msg NotificationMessage) {
					dedupedChan <- msg
					q.mux.Lock()
					delete(q.set, msg.User.ID+msg.WatchlistID)
					q.mux.Unlock()
					wg.Done()
				}(msg)
			}
		}
		wg.Wait()
		close(dedupedChan)
	}()

	return dedupedChan
}
