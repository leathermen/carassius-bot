package pkg

import (
	"errors"
	"fmt"
	"log"
	"time"

	"math/rand"

	"github.com/mymmrac/telego"
	"github.com/nikitades/carassius-bot/producer/pkg/router"
	"github.com/nikitades/carassius-bot/shared/request"
)

func (b *Bot) Handle(update telego.Update) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}

	if update.Message.Text == "/start" {
		b.start(update)
		return
	}

	mtype, err := b.router.Route(update.Message.Text)

	if err != nil {
		if errors.Is(err, router.ErrNoMedia) {
			if _, err = b.bot.SendMessage(&telego.SendMessageParams{
				ChatID: update.Message.Chat.ChatID(),
				Text:   NoLinkOrUnknownMedia,
			}); err != nil {
				log.Printf("failed to send unknown media msg, user %d, bot %s\n", update.Message.From.ID, b.name())
			}
		}
		return
	}

	exists, err := b.db.UserExistsInDB(update.Message.From.ID)
	if err != nil {
		log.Printf("failed to check the user for existence: %d\n", update.Message.From.ID)
	}

	if !exists {
		b.registerUser(update)
	} else {
		b.updateUser(update)
	}

	if err = b.db.AddUserMessageToDB(
		update.Message.From.ID,
		update.Message.From.FirstName,
		update.Message.From.LastName,
		update.Message.From.Username,
		update.Message.From.LanguageCode,
		update.Message.Text,
	); err != nil {
		log.Printf("failed to add user message to db, user %d, bot %s\n", update.Message.From.ID, b.name())
	}

	if mtype == request.TypeThanks {
		b.thanks(update)
		return
	}

	if err = b.publisher.AddMessageToQueue(update.Message.From.ID, update.Message.Text, b.name(), mtype.String()); err != nil {
		log.Printf("failed to add msg to queue, user %d, bot %s\n", update.Message.From.ID, b.name())
	}

	count, err := b.publisher.GetMessagesCountByBot(b.name())

	if err != nil {
		log.Println("failed to get remaining messages count")
		return
	}

	if _, err = b.bot.SendMessage(&telego.SendMessageParams{
		ChatID: update.Message.Chat.ChatID(),
		Text:   fmt.Sprintf("Don't block the bot. Wait. Your request is in the queue. There are currently %d messages ahead of yours.", count),
	}); err != nil {
		fmt.Printf("failed to send success msg, user %d, bot %s\n", update.Message.From.ID, b.name())
	}
}

func (b *Bot) start(update telego.Update) {
	if _, err := b.bot.SendMessage(&telego.SendMessageParams{
		ChatID: update.Message.Chat.ChatID(),
		Text:   StartMsg,
	}); err != nil {
		fmt.Printf("failed to send start msg, user %d, bot %s\n", update.Message.From.ID, b.name())
	}
}

func (b *Bot) thanks(update telego.Update) {
	var heartEmojis = []string{
		"\U0001F497", // 💗
		"\U00002764", // ❤️
		"\U0001F49B", // 💛
		"\U0001F499", // 💙
		"\U0001F49A", // 💚
		"\U0001F49C", // 💜
		"\U0001F495", // 💕
		"\U0001F496", // 💖
		"\U0001F49D", // 💝
		"\U0001F49E", // 💞
		"\U0001F49F", // 💟
		// Добавьте здесь другие эмоджи, если нужно
	}

	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(heartEmojis))
	heartEmoji := heartEmojis[index]

	if _, err := b.bot.SendMessage(&telego.SendMessageParams{
		ChatID: update.Message.Chat.ChatID(),
		Text:   heartEmoji,
	}); err != nil {
		fmt.Printf("failed to send thanks msg, user %d, bot %s\n", update.Message.From.ID, b.name())
	}
}

func (b *Bot) registerUser(update telego.Update) {
	err := b.db.AddUserToDB(*update.Message.From, b.name())
	if err != nil {
		log.Println("Error adding user to the database:", err)
		return
	}

	if _, err = b.bot.SendMessage(&telego.SendMessageParams{
		ChatID: update.Message.Chat.ChatID(),
		Text:   Hello,
	}); err != nil {
		fmt.Printf("failed to send hello msg, user %d, bot %s\n", update.Message.From.ID, b.name())
	}
}

func (b *Bot) updateUser(update telego.Update) {
	if err := b.db.UpdateUserInDB(*update.Message.From, b.name()); err != nil {
		log.Printf("failed to update a user in db, user %d, bot %s\n", update.Message.From.ID, b.name())
	}
}

func (b *Bot) name() string {
	botname, err := b.bot.GetMyName(nil)
	if err != nil {
		log.Fatal("failed to get own bot's name at registering the user")
	}

	return botname.Name
}
