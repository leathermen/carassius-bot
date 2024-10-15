package handlers

import (
	"github.com/mymmrac/telego"
	"github.com/nikitades/carassius-bot/consumer/internal/handlers/insta"
	"github.com/nikitades/carassius-bot/consumer/internal/handlers/pinterest"
	"github.com/nikitades/carassius-bot/consumer/internal/handlers/reddit"
	"github.com/nikitades/carassius-bot/consumer/internal/handlers/tiktok"
	"github.com/nikitades/carassius-bot/consumer/internal/handlers/twitter"
	"github.com/nikitades/carassius-bot/consumer/internal/handlers/youtube"
	"github.com/nikitades/carassius-bot/consumer/pkg/db"
	"github.com/nikitades/carassius-bot/consumer/pkg/queue"
	"github.com/nikitades/carassius-bot/shared/request"
)

type SuperHandler struct {
	handlers map[string]Handler
}

func NewSuper(bot *telego.Bot, q queue.Queue, db db.Database, channels []int64) *SuperHandler {
	handlers := make(map[string]Handler)

	instahandler := insta.New(bot, q, db, channels)
	handlers[instahandler.Name()] = instahandler

	ythandler := youtube.New(bot, q, db, channels)
	handlers[ythandler.Name()] = ythandler

	twhandler := twitter.New(bot, q, db, channels)
	handlers[twhandler.Name()] = twhandler

	reddithanadler := reddit.New(bot, q, db, channels)
	handlers[reddithanadler.Name()] = reddithanadler

	tiktokhandler := tiktok.New(bot, q)
	handlers[tiktokhandler.Name()] = tiktokhandler

	pinthandler := pinterest.New(bot, q)
	handlers[pinthandler.Name()] = pinthandler

	return &SuperHandler{
		handlers,
	}
}

func (sh *SuperHandler) HandleTiktok(userID int64, msg string, msgID int) {
	sh.handlers[request.TypeTiktok.String()].Handle(userID, msg, msgID)
}

func (sh *SuperHandler) HandleInsta(userID int64, msg string, msgID int) {
	sh.handlers[request.TypeInsta.String()].Handle(userID, msg, msgID)
}

func (sh *SuperHandler) HandleReddit(userID int64, msg string, msgID int) {
	sh.handlers[request.TypeReddit.String()].Handle(userID, msg, msgID)
}

func (sh *SuperHandler) HandleTwitter(userID int64, msg string, msgID int) {
	sh.handlers[request.TypeTwitter.String()].Handle(userID, msg, msgID)
}

func (sh *SuperHandler) HandleYoutube(userID int64, msg string, msgID int) {
	sh.handlers[request.TypeYoutube.String()].Handle(userID, msg, msgID)
}

func (sh *SuperHandler) HandlePinterest(userID int64, msg string, msgID int) {
	sh.handlers[request.TypePinterest.String()].Handle(userID, msg, msgID)
}
