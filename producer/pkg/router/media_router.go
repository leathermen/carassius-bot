package router

import (
	"errors"

	"github.com/nikitades/carassius-bot/shared/request"
)

var ErrNoMedia = errors.New("no media found")

type MediaRouter interface {
	Route(url string) (request.Type, error)
}
