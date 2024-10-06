package request

import (
	"errors"
	"log"
)

type Type uint

const (
	TypeTwitter Type = iota
	TypeInsta   Type = iota
	TypeReddit  Type = iota

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
	}

	log.Fatal(errors.ErrUnsupported)

	return "???"
}
