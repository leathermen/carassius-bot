package handler

import "errors"

var (
	ErrFailedToGetMedia = errors.New("failed to get media")
	ErrUnsupported      = errors.New("unsupported")
)

type Handler interface {
	HandleTiktok(userID int64, msg string, msgID int) error
	HandleInsta(userID int64, msg string, msgID int) error
	HandleReddit(userID int64, msg string, msgID int) error
	HandleTwitter(userID int64, msg string, msgID int) error
	HandleYoutube(userID int64, msg string, msgID int) error
	HandlePinterest(userID int64, msg string, msgID int) error
}
