package db

import "github.com/mymmrac/telego"

type Database interface {
	UserExistsInDB(userID int64) (bool, error)
	AddUserToDB(user telego.User, bot string) error
	UpdateUserInDB(user telego.User, bot string) error
	AddUserMessageToDB(userID int64, firstName, lastName, username, languageCode, message string) error
	AddMessageToQueue(userID int64, message, bot, socialNetworkName string) error
	GetMessagesCountByBot(botName string) (int, error)
}
