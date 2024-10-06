package db

import "github.com/nikitades/carassius-bot/consumer/pkg/queue"

type Database interface {
	GetMediaFileBySocialNetworkID(mediaID, platformName, botName string) (*queue.MediaFile, error)
	InsertMediaFile(mediaFile queue.MediaFile) error
}
