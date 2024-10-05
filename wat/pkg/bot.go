package pkg

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/mymmrac/telego"
	"github.com/nikitades/carassius-bot/producer/pkg/db"
	"github.com/nikitades/carassius-bot/producer/pkg/router"
)

type Bot struct {
	token  string
	bot    *telego.Bot
	router router.MediaRouter
	db     db.Database
}

func NewBot(
	config *BotConfig,
	mediaRouter router.MediaRouter,
	db db.Database,
) *Bot {
	return &Bot{
		token:  config.Token,
		router: mediaRouter,
		db:     db,
	}
}

func (b *Bot) Start(ctx context.Context) {
	var err error
	b.bot, err = telego.NewBot(b.token, telego.WithDefaultDebugLogger())

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	updates, _ := b.bot.UpdatesViaLongPolling(nil)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Bot's cancelled! Exiting...")
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
