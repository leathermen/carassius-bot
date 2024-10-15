package pinterest

import (
	"log"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
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

func (ph *Handler) Handle(userID int64, _ string, msgID int) {
	defer func() {
		if err := ph.q.DeleteMessageFromQueue(msgID); err != nil {
			log.Printf("failed to remove message from queue: %d", msgID)
		}
	}()

	if _, err := ph.bot.SendMessage(&telego.SendMessageParams{
		ChatID: telegoutil.ID(userID),
		Text:   "Sorry, this media platform will be supported only in 2025!",
	}); err != nil {
		log.Printf("failed to send platform is not supported yet message, user %d", userID)
	}
}

func (ph *Handler) Name() string {
	return Code
}
