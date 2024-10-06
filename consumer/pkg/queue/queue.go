package queue

type Queue interface {
	GetMessageFromQueueByBot(botName string) (*Message, error)
	DeleteMessageFromQueue(messageID int) error
}
