package request

import (
	"errors"
	"log"
)

type Type uint

const (
	TypeTwitter   Type = iota
	TypeInsta     Type = iota
	TypeReddit    Type = iota
	TypeTiktok    Type = iota
	TypeYoutube   Type = iota
	TypePinterest Type = iota

	TypeThanks Type = iota //uWu
)

func (rt Type) String() string {
	switch rt {
	case TypeTwitter:
		return "twitter"
	case TypeInsta:
		return "instagram"
	case TypeReddit:
		return "reddit"
	case TypeTiktok:
		return "tiktok"
	case TypeYoutube:
		return "youtube"
	case TypePinterest:
		return "pinterest"
	}

	log.Fatal(errors.ErrUnsupported)

	return "???"
}
