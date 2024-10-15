package twitter

type TweetType int

const (
	Video TweetType = iota
	Gif   TweetType = iota
	Photo TweetType = iota
)
