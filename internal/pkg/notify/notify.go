package notify

type Notifier interface {
	Send() error
}
