package youtube

import (
	"fmt"

	"github.com/mymmrac/telego"
	"github.com/nikitades/carassius-bot/consumer/pkg/queue"
)

const Code = "youtube"

type handler struct {
	bot *telego.Bot
	q   queue.Queue
}

func New(bot *telego.Bot, q queue.Queue) *handler {
	return &handler{bot, q}
}

func (h *handler) Handle(userID int64, msg string, msgID int) {
	fmt.Printf("userID: %d, msg: %s, msgID: %d", userID, msg, msgID)

	h.q.DeleteMessageFromQueue(msgID)
}

func (h *handler) Name() string {
	return Code
}
