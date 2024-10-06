package queue

import "time"

type MediaFile struct {
	ID                int
	SocialNetworkID   string
	SocialNetworkName string
	FileID            string
	FileType          string
	Bot               string
}

type Message struct {
	ID                int
	UserID            int64
	Message           string
	BotName           string
	SocialNetworkName string
	Timestamp         time.Time
}
