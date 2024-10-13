package db

import (
	"database/sql"
	"log"
	"regexp"

	"github.com/mymmrac/telego"
)

type Database struct {
	db *sql.DB
}

func New(db *sql.DB) *Database {
	return &Database{db}
}

func (d *Database) UserExistsInDB(userID int64) (bool, error) {
	query := "SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)"

	var exists bool

	err := d.db.QueryRow(query, userID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (d *Database) AddUserToDB(user telego.User, bot string) error {
	// SQL-запрос для вставки данных пользователя в таблицу users
	query := `
        INSERT INTO users (id, is_bot, first_name, last_name, username, language_code, can_join_groups, can_read_all_group_messages, supports_inline_queries, bot, update)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
    `

	// Выполняем SQL-запрос
	_, err := d.db.Exec(query, user.ID, user.IsBot, user.FirstName, user.LastName, user.Username, user.LanguageCode, user.CanJoinGroups, user.CanReadAllGroupMessages, user.SupportsInlineQueries, bot)

	if err != nil {
		return err
	}

	return nil
}

func (d *Database) UpdateUserInDB(user telego.User, bot string) error {
	query := `
        UPDATE users
        SET
            is_bot = $2,
            first_name = $3,
            last_name = $4,
            username = $5,
            language_code = $6,
            can_join_groups = $7,
            can_read_all_group_messages = $8,
            supports_inline_queries = $9,
            bot = $10,
            update = NOW()
        WHERE id = $1
    `

	_, err := d.db.Exec(query, user.ID, user.IsBot, user.FirstName, user.LastName, user.Username, user.LanguageCode, user.CanJoinGroups, user.CanReadAllGroupMessages, user.SupportsInlineQueries, bot)

	return err
}

func (d *Database) AddUserMessageToDB(userID int64, firstName, lastName, username, languageCode, message string) error {
	// SQL-запрос для вставки сообщения пользователя в таблицу user_message
	query := `
        INSERT INTO user_message (user_id, first_name, last_name, username, language_code, message)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

	// Выполняем SQL-запрос
	_, err := d.db.Exec(query, userID, firstName, lastName, username, languageCode, message)

	if err != nil {
		return err
	}

	return nil
}

// AddMessageToQueue добавляет ссылки TikTok из сообщения в очередь базы данных.
func (d *Database) AddMessageToQueue(userID int64, message, bot, socialNetworkName string) error {
	// Регулярное выражение для поиска ссылок в тексте сообщения
	linkRegex := regexp.MustCompile(`https:\/\/(?:www\.)?([a-zA-Z0-9-]+(\.[a-zA-Z]{2,})+)(\/\S*)?`)

	// Ищем ссылки в тексте сообщения
	matches := linkRegex.FindAllString(message, -1)

	// Если найдены ссылки, добавляем уникальные в базу данных для конкретного пользователя
	if len(matches) > 0 { //nolint:nestif
		for _, link := range matches {
			// Проверяем, есть ли такая ссылка уже в базе для данного пользователя
			var count int
			err := d.db.QueryRow("SELECT COUNT(*) FROM message_queue WHERE user_id = $1 AND message = $2", userID, link).Scan(&count)
			if err != nil {
				return err
			}

			// Если ссылки нет в базе для данного пользователя, добавляем ее
			if count == 0 {
				_, err := d.db.Exec(
					"INSERT INTO message_queue (user_id, message, bot, social_network_name) VALUES ($1, $2, $3, $4)",
					userID, link, bot, socialNetworkName,
				)
				if err != nil {
					return err
				}
			}
		}
	} else {
		// Можно отправить пользователю сообщение о том, что не найдено ссылок
		log.Println("No links found in the message.")
	}

	return nil
}

func (d *Database) GetMessagesCountByBot(botName string) (int, error) {
	var count int
	err := d.db.QueryRow(
		"SELECT COUNT(*)-1 FROM message_queue WHERE bot = $1",
		botName,
	).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}
