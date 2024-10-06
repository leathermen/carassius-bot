package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/nikitades/carassius-bot/consumer/pkg/queue"
)

type Database struct {
	db *sql.DB
}

func New(db *sql.DB) *Database {
	return &Database{db}
}

func (d *Database) GetMessageFromQueueByBot(botName string) (*queue.Message, error) {
	query := "SELECT id, user_id, message, bot, social_network_name, timestamp FROM message_queue WHERE bot = $1 ORDER BY timestamp LIMIT 1"

	rows, err := d.db.Query(query, botName) //nolint:rowserrcheck
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

	return nil, nil //nolint:nilnil
}

func (d *Database) DeleteMessageFromQueue(messageID int) error {
	query := "DELETE FROM message_queue WHERE id = $1"
	_, err := d.db.Exec(query, messageID)
	return err
}

func (d *Database) GetMediaFileBySocialNetworkID(mediaID, platformName, botName string) (*queue.MediaFile, error) {
	query := "SELECT id, social_network_id, social_network_name, file_id, file_type FROM media_files WHERE social_network_id = $1 AND social_network_name = $2 AND bot = $3"
	row := d.db.QueryRow(query, mediaID, platformName, botName)

	var mediaFile queue.MediaFile
	err := row.Scan(&mediaFile.ID, &mediaFile.SocialNetworkID, &mediaFile.SocialNetworkName, &mediaFile.FileID, &mediaFile.FileType)
	if err == sql.ErrNoRows {
		return nil, nil //nolint:nilnil
	} else if err != nil {
		return nil, err
	}

	return &mediaFile, nil
}

func (d *Database) InsertMediaFile(mediaFile queue.MediaFile) error {
	query := `
		INSERT INTO media_files (social_network_id, social_network_name, file_id, file_type, bot)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := d.db.QueryRow(query, mediaFile.SocialNetworkID, mediaFile.SocialNetworkName, mediaFile.FileID, mediaFile.FileType, mediaFile.Bot).Scan(&mediaFile.ID)
	if err != nil {
		log.Println("Error inserting media file:", err)
		return err
	}

	fmt.Printf("Inserted media file with ID %d\n", mediaFile.ID)
	return nil
}
