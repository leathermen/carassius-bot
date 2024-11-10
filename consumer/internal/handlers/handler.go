package handlers

type Handler interface {
	Handle(userID int64, msg string, msgID int) error
	Name() string
}
