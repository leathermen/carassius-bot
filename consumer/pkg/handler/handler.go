package handler

type Handler interface {
	HandleTiktok(userID int64, msg string, msgID int)
	HandleInsta(userID int64, msg string, msgID int)
	HandleReddit(userID int64, msg string, msgID int)
	HandleTwitter(userID int64, msg string, msgID int)
	HandleYoutube(userID int64, msg string, msgID int)
	HandlePinterest(userID int64, msg string, msgID int)
}
