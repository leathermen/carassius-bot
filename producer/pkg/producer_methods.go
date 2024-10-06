package pkg

import (
	"log"
	"time"

	"math/rand"

	"github.com/mymmrac/telego"
	"github.com/nikitades/carassius-bot/shared/bothelper"
)

func (b *Producer) start(update telego.Update) {
	if _, err := b.bot.SendMessage(&telego.SendMessageParams{
		ChatID: update.Message.Chat.ChatID(),
		Text:   StartMsg,
	}); err != nil {
		log.Printf("failed to send start msg, user %d, bot %s: %s", update.Message.From.ID, bothelper.Botname(b.bot), err)
	}
}

func (b *Producer) thanks(update telego.Update) {
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
		log.Printf("failed to send thanks msg, user %d, bot %s: %s", update.Message.From.ID, bothelper.Botname(b.bot), err)
	}
}

func (b *Producer) registerUser(update telego.Update) {
	err := b.db.AddUserToDB(*update.Message.From, bothelper.Botname(b.bot))
	if err != nil {
		log.Println("Error adding user to the database:", err)
		return
	}

	if _, err = b.bot.SendMessage(&telego.SendMessageParams{
		ChatID: update.Message.Chat.ChatID(),
		Text:   Hello,
	}); err != nil {
		log.Printf("failed to send hello msg, user %d, bot %s: %s", update.Message.From.ID, bothelper.Botname(b.bot), err)
	}
}

func (b *Producer) updateUser(update telego.Update) {
	if err := b.db.UpdateUserInDB(*update.Message.From, bothelper.Botname(b.bot)); err != nil {
		log.Printf("failed to update a user in db, user %d, bot %s: %s", update.Message.From.ID, bothelper.Botname(b.bot), err)
	}
}
