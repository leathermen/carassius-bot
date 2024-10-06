package main

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mymmrac/telego"
	"github.com/nikitades/carassius-bot/producer/internal/db"
	"github.com/nikitades/carassius-bot/producer/internal/router"
	"github.com/nikitades/carassius-bot/producer/pkg"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	dbconn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to create db connection: %s", err)
	}

	db := db.New(dbconn)
	mediaRouter := router.New()

	botToken := os.Getenv("TOKEN")
	bot, err := telego.NewBot(botToken, telego.WithDefaultLogger(false, true))

	producer := pkg.NewProducer(bot, mediaRouter, db, db)

	var signalChan chan (os.Signal) = make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-signalChan
		cancel()
	}()

	producer.Start(ctx)
}
