package insta

import "log"

type PostType int

const (
	Post  PostType = iota
	Story PostType = iota
	Reel  PostType = iota
)

func (pt PostType) String() string {
	switch pt {
	case Post:
		return "post"
	case Story:
		return "story"
	case Reel:
		return "reel"
	default:
		log.Fatalf("unexpected instagram entity type: %d", pt)
		return ""
	}
}
