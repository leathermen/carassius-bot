package insta

import (
	"github.com/mymmrac/telego"
	"github.com/nikitades/carassius-bot/consumer/pkg/queue"
)

const Code = "insta"

type Handler struct {
	bot *telego.Bot
	q   queue.Queue
}

func New(bot *telego.Bot, q queue.Queue) *Handler {
	return &Handler{bot, q}
}

func (h *Handler) Handle(_ int64, _ string, _ int) {

}

func (h *Handler) Name() string {
	return Code
}
