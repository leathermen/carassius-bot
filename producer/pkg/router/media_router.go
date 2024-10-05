package router

import (
	"errors"
	"log"
)

var ErrNoMedia = errors.New("no media found")

type RequestType uint

const (
	RequestTypeTwitter RequestType = iota
	RequestTypeInsta   RequestType = iota
	RequestTypeReddit  RequestType = iota

	RequestTypeThanks RequestType = iota //uWu
)

func (rt RequestType) String() string {
	switch rt {
	case RequestTypeTwitter:
		return "twitter"
	case RequestTypeInsta:
		return "instagram"
	case RequestTypeReddit:
		return "reddit"
	}

	log.Fatal(errors.ErrUnsupported)

	return "???"
}

type MediaRouter interface {
	Route(url string) (RequestType, error)
}
