package insta

import (
	"github.com/mymmrac/telego"
	"github.com/nikitades/carassius-bot/consumer/pkg/queue"
)

const Code = "insta"

type handler struct {
	bot *telego.Bot
	q   queue.Queue
}

func New(bot *telego.Bot, q queue.Queue) *handler {
	return &handler{bot, q}
}

func (h *handler) Handle(userID int64, msg string, msgID int) {

}

func (h *handler) Name() string {
	return Code
}
