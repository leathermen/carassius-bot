package handlers

import (
	"github.com/mymmrac/telego"
	"github.com/nikitades/carassius-bot/consumer/internal/handlers/insta"
	"github.com/nikitades/carassius-bot/consumer/internal/handlers/youtube"
	"github.com/nikitades/carassius-bot/consumer/pkg/db"
	"github.com/nikitades/carassius-bot/consumer/pkg/queue"
)

type SuperHandler struct {
	handlers map[string]Handler
}

func NewSuper(bot *telego.Bot, q queue.Queue, db db.Database) *SuperHandler {
	handlers := make(map[string]Handler)

	instahandler := insta.New(bot, q, db)
	handlers[instahandler.Name()] = instahandler

	ythandler := youtube.New(bot, q, db)
	handlers[ythandler.Name()] = ythandler

	return &SuperHandler{
		handlers,
	}
}

func (sh *SuperHandler) HandleTiktok(userID int64, msg string, msgID int) {
	sh.handlers["tiktok"].Handle(userID, msg, msgID)
}

func (sh *SuperHandler) HandleInsta(userID int64, msg string, msgID int) {
	sh.handlers["insta"].Handle(userID, msg, msgID)
}

func (sh *SuperHandler) HandleReddit(userID int64, msg string, msgID int) {
	sh.handlers["reddit"].Handle(userID, msg, msgID)
}

func (sh *SuperHandler) HandleTwitter(userID int64, msg string, msgID int) {
	sh.handlers["twitter"].Handle(userID, msg, msgID)
}

func (sh *SuperHandler) HandleYoutube(userID int64, msg string, msgID int) {
	sh.handlers["youtube"].Handle(userID, msg, msgID)
}

func (sh *SuperHandler) HandlePinterest(userID int64, msg string, msgID int) {
	sh.handlers["pinterest"].Handle(userID, msg, msgID)
}
