package handler

import (
	"errors"
	"fmt"

	"github.com/thanhpk/randstr"
)

var (
	ErrFailedToGetMedia = errors.New("failed to get media")
	ErrUnsupported      = errors.New("unsupported")
)

type Handler interface {
	HandleTiktok(userID int64, msg string, msgID int) error
	HandleInsta(userID int64, msg string, msgID int) error
	HandleReddit(userID int64, msg string, msgID int) error
	HandleTwitter(userID int64, msg string, msgID int) error
	HandleYoutube(userID int64, msg string, msgID int) error
	HandlePinterest(userID int64, msg string, msgID int) error
}

type Proxy struct {
	Username   string
	CountryISO string
	SID        string
	Password   string
	Hostname   string
	Port       string
}

func (p *Proxy) UsernameWithCountry() string {
	return fmt.Sprintf(p.Username, p.CountryISO)
}

func (p *Proxy) UsernameWithCountrySID() string {
	return fmt.Sprintf(p.Username, p.CountryISO, p.SID)
}

func (p *Proxy) UsernameWithCountryAndRandomSID() string {
	return fmt.Sprintf(p.Username, p.CountryISO, randstr.Base62(16))
}

func (p *Proxy) HostnamePort() string {
	return p.Hostname + ":" + p.Port
}
