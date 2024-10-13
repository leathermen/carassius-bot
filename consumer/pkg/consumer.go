package pkg

import (
	"context"
	"log"
	"time"

	"github.com/nikitades/carassius-bot/consumer/pkg/handler"
	"github.com/nikitades/carassius-bot/consumer/pkg/queue"
	"github.com/nikitades/carassius-bot/shared/request"
	"golang.org/x/time/rate"
)

type Consumer struct {
	botname string
	queue   queue.Queue
	handler handler.Handler
}

func NewConsumer(botname string, queue queue.Queue, handler handler.Handler) *Consumer {
	return &Consumer{
		botname: botname,
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
			msg, err := c.queue.GetMessageFromQueueByBot(c.botname)

			if err != nil {
				log.Printf("failed getting update from queue, bot name %s: %s", c.botname, err)
			}

			if msg == nil {
				continue
			}

			switch msg.SocialNetworkName {
			case request.TypeTiktok.String():
				c.handler.HandleTiktok(msg.UserID, msg.Message, msg.ID)
			case request.TypeInsta.String():
				c.handler.HandleInsta(msg.UserID, msg.Message, msg.ID)
			case request.TypeTwitter.String():
				c.handler.HandleTwitter(msg.UserID, msg.Message, msg.ID)
			case request.TypeYoutube.String():
				c.handler.HandleYoutube(msg.UserID, msg.Message, msg.ID)
			case request.TypePinterest.String():
				c.handler.HandlePinterest(msg.UserID, msg.Message, msg.ID)
			case request.TypeReddit.String():
				c.handler.HandleReddit(msg.UserID, msg.Message, msg.ID)
			default:
				log.Printf("Unknown message type: %s", msg.SocialNetworkName)
				if err := c.queue.DeleteMessageFromQueue(msg.ID); err != nil {
					log.Printf("failed to remove msg from queue: %d", msg.ID)
				}
			}
		}
	}
}
