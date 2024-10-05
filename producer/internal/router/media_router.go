package router

import (
	"net/url"
	"regexp"
	"strings"

	pkgrouter "github.com/nikitades/carassius-bot/producer/pkg/router"
)

type MediaRouter struct {
}

func New() *MediaRouter {
	return &MediaRouter{}
}

func (mr *MediaRouter) Route(text string) (pkgrouter.RequestType, error) {
	if mr.tryTwitter(text) {
		return pkgrouter.RequestTypeTwitter, nil
	}

	if mr.tryInsta(text) {
		return pkgrouter.RequestTypeInsta, nil
	}

	if mr.tryReddit(text) {
		return pkgrouter.RequestTypeReddit, nil
	}

	if mr.tryThanks(text) {
		return pkgrouter.RequestTypeThanks, nil
	}

	return 0, pkgrouter.ErrNoMedia
}

func (mr *MediaRouter) tryTwitter(text string) bool {
	return mr.try(text, "x.com")
}

func (mr *MediaRouter) tryInsta(text string) bool {
	return mr.try(text, "instagram.com")
}

func (mr *MediaRouter) tryReddit(text string) bool {
	return mr.try(text, "reddit.com")
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

func (mr *MediaRouter) tryThanks(text string) bool {
	// Создаем регулярное выражение для поиска ключевых фраз
	regex := regexp.MustCompile(`(?i)\b(thanks|thank you|thx|tq|thanks a lot|thanks a bunch|❤)\b`)

	// Используем FindStringSubmatch для поиска соответствий
	matches := regex.FindStringSubmatch(text)

	// Если найдено соответствие, возвращаем true
	return len(matches) > 0
}
