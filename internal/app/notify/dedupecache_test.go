package notify

import (
	"fmt"
	"testing"

	"github.com/scirelli/auction-ebidlocal-search/internal/app/server/model"
	"github.com/stretchr/testify/assert"
)

func fixture_GenerateMsgs(numMsgs int, numDupes int) <-chan NotificationMessage {
	var msgChan = make(chan NotificationMessage)
	go func() {
		for i, dupes := 0, 0; i < numMsgs; i++ {
			v := fmt.Sprintf("%d", i)
			fmt.Printf("Writing %s\n", v)
			msgChan <- NotificationMessage{
				User: &model.User{
					Name: v,
					ID:   v,
				},
				WatchlistID: v,
			}
			if dupes < numDupes {
				msgChan <- NotificationMessage{
					User: &model.User{
						Name: v,
						ID:   v,
					},
					WatchlistID: v,
				}
			}
		}
		close(msgChan)
	}()
	return msgChan
}

func fixture_GenerateMsgsDupesLater(numMsgs int, numDupes int) <-chan NotificationMessage {
	var msgChan = make(chan NotificationMessage)
	go func() {
		for i := 0; i < numMsgs; i++ {
			v := fmt.Sprintf("%d", i)
			fmt.Printf("Writing %s\n", v)
			msgChan <- NotificationMessage{
				User: &model.User{
					Name: v,
					ID:   v,
				},
				WatchlistID: v,
			}
		}
		for i := 0; i < numDupes; i++ {
			v := fmt.Sprintf("%d", i)
			fmt.Printf("Writing %s\n", v)
			msgChan <- NotificationMessage{
				User: &model.User{
					Name: v,
					ID:   v,
				},
				WatchlistID: v,
			}
		}
		close(msgChan)
	}()
	return msgChan
}

func Test_DedupeQueueNoDupes(t *testing.T) {
	subject := NewDedupeQueue()
	expected := 10
	dqChan := subject.Enqueue(fixture_GenerateMsgs(expected, 0))

	count := 0
	for msg := range dqChan {
		fmt.Printf("Actual: %s\n", msg.WatchlistID)
		count++
	}
	assert.Equalf(t, expected, count, "Expected: %s, actual: %s", expected, count)
}

func Test_DedupeQueueTwoDupes(t *testing.T) {
	subject := NewDedupeQueue()
	expected := 10
	dqChan := subject.Enqueue(fixture_GenerateMsgs(expected, 2))

	count := 0
	for msg := range dqChan {
		fmt.Printf("Actual: %s\n", msg.WatchlistID)
		count++
	}
	assert.Equalf(t, expected, count, "Expected: %s, actual: %s", expected, count)
}

func Test_DedupeQueueTenDupes(t *testing.T) {
	subject := NewDedupeQueue()
	expected := 10
	dqChan := subject.Enqueue(fixture_GenerateMsgs(expected, 10))

	count := 0
	for msg := range dqChan {
		fmt.Printf("Actual: %s\n", msg.WatchlistID)
		count++
	}
	assert.Equalf(t, expected, count, "Expected: %s, actual: %s", expected, count)
}

func Test_DedupeQueueMixDupes(t *testing.T) {
	subject := NewDedupeQueue()
	expected := 15
	dqChan := subject.Enqueue(fixture_GenerateMsgsDupesLater(10, 5))

	count := 0
	for msg := range dqChan {
		fmt.Printf("Actual: %s\n", msg.WatchlistID)
		count++
	}
	assert.Equalf(t, expected, count, "Expected: %s, actual: %s", expected, count)
}
