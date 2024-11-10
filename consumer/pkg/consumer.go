package pkg

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/nikitades/carassius-bot/consumer/pkg/handler"
	"github.com/nikitades/carassius-bot/consumer/pkg/queue"
	"github.com/nikitades/carassius-bot/shared/bothelper"
	"github.com/nikitades/carassius-bot/shared/request"
	"golang.org/x/time/rate"
)

type Consumer struct {
	bot     *telego.Bot
	queue   queue.Queue
	handler handler.Handler
}

func NewConsumer(bot *telego.Bot, queue queue.Queue, handler handler.Handler) *Consumer {
	return &Consumer{
		bot:     bot,
		queue:   queue,
		handler: handler,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	limiter := rate.NewLimiter(rate.Every(time.Second*5), 1)
	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer's stopped! Exiting...")
			return
		default:
			_ = limiter.Wait(ctx)
			botname := bothelper.Botname(c.bot)
			msg, err := c.queue.GetMessageFromQueueByBot(botname)

			if err != nil {
				log.Printf("failed getting update from queue, bot name %s: %s", botname, err)
			}

			if msg == nil {
				continue
			}

			switch msg.SocialNetworkName {
			case request.TypeTiktok.String():
				err = c.handler.HandleTiktok(msg.UserID, msg.Message, msg.ID)
			case request.TypeInsta.String():
				err = c.handler.HandleInsta(msg.UserID, msg.Message, msg.ID)
			case request.TypeTwitter.String():
				err = c.handler.HandleTwitter(msg.UserID, msg.Message, msg.ID)
			case request.TypeYoutube.String():
				err = c.handler.HandleYoutube(msg.UserID, msg.Message, msg.ID)
			case request.TypePinterest.String():
				err = c.handler.HandlePinterest(msg.UserID, msg.Message, msg.ID)
			case request.TypeReddit.String():
				err = c.handler.HandleReddit(msg.UserID, msg.Message, msg.ID)
			default:
				log.Printf("Unknown message type: %s", msg.SocialNetworkName)
				if err := c.queue.DeleteMessageFromQueue(msg.ID); err != nil {
					log.Printf("failed to remove msg from queue: %d", msg.ID)
				}
			}

			if errors.Is(err, handler.ErrUnsupported) {
				if _, err := c.bot.SendMessage(&telego.SendMessageParams{
					ChatID: telegoutil.ID(msg.UserID),
					Text:   "This type of media is not supported. Supported: reels, posts, pictures, videos.",
				}); err != nil {
					log.Printf("failed to send unsupported message")
				}
			}

			if errors.Is(err, handler.ErrFailedToGetMedia) {
				if _, err := c.bot.SendMessage(&telego.SendMessageParams{
					ChatID: telegoutil.ID(msg.UserID),
					Text:   "Failed to get media!",
				}); err != nil {
					log.Printf("failed to send failed to find media message, user %d", msg.UserID)
				}
			}
		}
	}
}
