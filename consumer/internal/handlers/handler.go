package handlers

type handler interface {
	Handle(userID int64, msg string, msgID int)
	Name() string
}
