package db

import (
	"database/sql"

	"github.com/nikitades/carassius-bot/consumer/pkg/queue"

	_ "github.com/lib/pq"
)

type Database struct {
	db *sql.DB
}

func New(db *sql.DB) *Database {
	return &Database{db}
}

func (d *Database) GetMessageFromQueueByBot(botName string) (*queue.Message, error) {
	query := "SELECT id, user_id, message, bot, social_network_name, timestamp FROM message_queue WHERE bot = $1 ORDER BY timestamp LIMIT 1"

	rows, err := d.db.Query(query, botName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var message queue.Message
		err := rows.Scan(&message.ID, &message.UserID, &message.Message, &message.BotName, &message.SocialNetworkName, &message.Timestamp)
		if err != nil {
			return nil, err
		}
		return &message, nil
	}

	return nil, nil
}

func (d *Database) DeleteMessageFromQueue(messageID int) error {
	query := "DELETE FROM message_queue WHERE id = $1"
	_, err := d.db.Exec(query, messageID)
	return err
}
