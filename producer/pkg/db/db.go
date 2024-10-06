package db

import "github.com/mymmrac/telego"

type Database interface {
	UserExistsInDB(userID int64) (bool, error)
	AddUserToDB(user telego.User, bot string) error
	UpdateUserInDB(user telego.User, bot string) error
	AddUserMessageToDB(userID int64, firstName, lastName, username, languageCode, message string) error
}
