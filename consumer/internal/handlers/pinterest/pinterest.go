package pinterest

import (
	"log"

	"github.com/mymmrac/telego"
	"github.com/nikitades/carassius-bot/consumer/pkg/handler"
	"github.com/nikitades/carassius-bot/consumer/pkg/queue"
)

const Code = "pinterest"

type Handler struct {
	bot *telego.Bot
	q   queue.Queue
}

func New(bot *telego.Bot, q queue.Queue) *Handler {
	return &Handler{bot, q}
}

func (ph *Handler) Handle(_ int64, _ string, msgID int) error {
	defer func() {
		if err := ph.q.DeleteMessageFromQueue(msgID); err != nil {
			log.Printf("failed to remove message from queue: %d", msgID)
		}
	}()

	return handler.ErrUnsupported
}

func (ph *Handler) Name() string {
	return Code
}
