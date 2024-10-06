package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mymmrac/telego"
	"github.com/nikitades/carassius-bot/consumer/internal/db"
	"github.com/nikitades/carassius-bot/consumer/internal/handlers"
	"github.com/nikitades/carassius-bot/consumer/pkg"
	"github.com/nikitades/carassius-bot/shared/bothelper"

	_ "github.com/lib/pq"
)

func main() {
	botToken := os.Getenv("TOKEN")

	dbURL := os.Getenv("DATABASE_URL")
	dbconn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to create db connection: %s", err)
	}

	db := db.New(dbconn)

	debug := os.Getenv("DEBUG") == "1"

	var loggerOption telego.BotOption

	if debug {
		loggerOption = telego.WithDefaultDebugLogger()
	} else {
		loggerOption = telego.WithDefaultLogger(false, true)
	}

	bot, err := telego.NewBot(botToken, loggerOption)
	if err != nil {
		log.Fatalf("failed to create tg bot: %s", err)
	}

	consumer := pkg.NewConsumer(bothelper.Botname(bot), db, handlers.NewSuper(bot, db, db))

	var signalChan chan (os.Signal) = make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-signalChan
		cancel()
	}()

	consumer.Start(ctx)
}
