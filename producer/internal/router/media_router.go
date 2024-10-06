package router

import (
	"net/url"
	"strings"

	pkgrouter "github.com/nikitades/carassius-bot/producer/pkg/router"
	"github.com/nikitades/carassius-bot/shared/request"
)

type MediaRouter struct{}

func New() *MediaRouter {
	return &MediaRouter{}
}

func (mr *MediaRouter) Route(text string) (request.Type, error) {
	if mr.tryTwitter(text) {
		return request.TypeTwitter, nil
	}

	if mr.tryInsta(text) {
		return request.TypeInsta, nil
	}

	if mr.tryReddit(text) {
		return request.TypeReddit, nil
	}

	if mr.tryTiktok(text) {
		return request.TypeTiktok, nil
	}

	if mr.tryPinterest(text) {
		return request.TypePinterest, nil
	}

	if mr.tryYoutube(text) {
		return request.TypeYoutube, nil
	}

	if mr.tryThanks(text) {
		return request.TypeThanks, nil
	}

	return 0, pkgrouter.ErrNoMedia
}

func (mr *MediaRouter) tryTwitter(text string) bool {
	return mr.try(text, "x.com") || mr.try(text, "twitter.com")
}

func (mr *MediaRouter) tryInsta(text string) bool {
	return mr.try(text, "instagram.com")
}

func (mr *MediaRouter) tryReddit(text string) bool {
	return mr.try(text, "reddit.com")
}

func (mr *MediaRouter) tryPinterest(text string) bool {
	return mr.try(text, "pinterest.com") || mr.try(text, "pin.it")
}

func (mr *MediaRouter) tryYoutube(text string) bool {
	return mr.try(text, "youtube.com") || mr.try(text, "youtu.be")
}

func (mr *MediaRouter) tryTiktok(text string) bool {
	return mr.try(text, "tiktok.com")
}

func (mr *MediaRouter) try(text, host string) bool {
	url, err := url.Parse(text)

	if err != nil {
		return false
	}

	if strings.Contains(url.Host, host) {
		return true
	}

	return false
}
