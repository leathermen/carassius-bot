package pkg

import (
	"context"
	"log"
	"sync"

	"github.com/mymmrac/telego"
	"github.com/nikitades/carassius-bot/producer/pkg/db"
	"github.com/nikitades/carassius-bot/producer/pkg/publisher"
	"github.com/nikitades/carassius-bot/producer/pkg/router"
)

type Producer struct {
	bot       *telego.Bot
	router    router.MediaRouter
	db        db.Database
	publisher publisher.Publisher
}

func NewProducer(
	bot *telego.Bot,
	mediaRouter router.MediaRouter,
	db db.Database,
	publisher publisher.Publisher,
) *Producer {
	return &Producer{
		bot:       bot,
		router:    mediaRouter,
		db:        db,
		publisher: publisher,
	}
}

func (b *Producer) Start(ctx context.Context) {
	updates, _ := b.bot.UpdatesViaLongPolling(nil)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		log.Println("Producer started...")
		for {
			select {
			case <-ctx.Done():
				log.Println("Bot's cancelled! Exiting...")
				wg.Done()
				return
			case update := <-updates:
				b.Handle(update)
			}
		}
	}()

	<-ctx.Done()
	b.bot.StopLongPolling()
	wg.Wait()
}
