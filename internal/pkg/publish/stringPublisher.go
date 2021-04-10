package publish

type StringPublisher interface {
	Subscribe() (<-chan string, func() error)
	Publish(string)
}
