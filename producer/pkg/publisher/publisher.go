package publisher

type Publisher interface {
	AddMessageToQueue(userID int64, message, bot, socialNetworkName string) error
	GetMessagesCountByBot(botName string) (int, error)
}
