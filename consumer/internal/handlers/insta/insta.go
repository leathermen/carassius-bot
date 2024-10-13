package insta

import (
	"log"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/nikitades/carassius-bot/consumer/pkg/db"
	"github.com/nikitades/carassius-bot/consumer/pkg/queue"
)

const Code = "insta"

type subhandler interface {
	Handle(userID int64, msg string, msgID int)
}

type Handler struct {
	bot      *telego.Bot
	q        queue.Queue
	handlers map[PostType]subhandler
}

func New(bot *telego.Bot, q queue.Queue, db db.Database) *Handler {
	csrfprovider := newCsrfProvider()

	handlers := map[PostType]subhandler{}
	handlers[Reel] = newReelHandler(bot, db, csrfprovider)

	return &Handler{bot, q, handlers}
}

func (h *Handler) Name() string {
	return Code
}

func (h *Handler) Handle(userID int64, msg string, msgID int) {
	defer func() {
		if err := h.q.DeleteMessageFromQueue(msgID); err != nil {
			log.Printf("failed to remove message from queue: %d", msgID)
		}
	}()

	postType, found := getPostType(msg)

	if !found {
		log.Printf("unsupported insta media provided")
		if _, err := h.bot.SendMessage(&telego.SendMessageParams{
			ChatID: telegoutil.ID(userID),
			Text:   "This type of media is not supported.",
		}); err != nil {
			log.Printf("failed to send unsupported message")
		}

		return
	}

	h.handlers[postType].Handle(userID, msg, msgID)
}
