package tiktok

import (
	"log"

	"github.com/mymmrac/telego"
	"github.com/nikitades/carassius-bot/consumer/pkg/handler"
	"github.com/nikitades/carassius-bot/consumer/pkg/queue"
)

const Code = "tiktok"

type Handler struct {
	bot *telego.Bot
	q   queue.Queue
}

func New(bot *telego.Bot, q queue.Queue) *Handler {
	return &Handler{bot, q}
}

func (th *Handler) Handle(_ int64, _ string, msgID int) error {
	defer func() {
		if err := th.q.DeleteMessageFromQueue(msgID); err != nil {
			log.Printf("failed to remove message from queue: %d", msgID)
		}
	}()

	return handler.ErrUnsupported
}

func (th *Handler) Name() string {
	return Code
}
