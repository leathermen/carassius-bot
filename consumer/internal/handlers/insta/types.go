package insta

type PostType int

const (
	Post  PostType = iota
	Story PostType = iota
	Reel  PostType = iota
)
