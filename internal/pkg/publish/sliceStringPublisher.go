package publish

type SliceStringPublisher interface {
	Register() (<-chan []string, func() error)
	Publish([]string)
}
