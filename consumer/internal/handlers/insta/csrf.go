package insta

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	ErrNotFound = errors.New("CSRF cookie not found")
)

type csrfprovider struct {
	csrf     string
	csrfLock sync.RWMutex
}

func newCsrfProvider() *csrfprovider {
	cp := &csrfprovider{}
	cp.csrf, _ = fetchCSRF()

	go func() {
		timer := time.NewTicker(time.Minute * 15)
		for {
			<-timer.C
			newCSRF, err := fetchCSRF()
			if err != nil {
				log.Printf("failed to background fetch CSRF: %s", err)
				continue
			}

			cp.csrfLock.Lock()
			cp.csrf = newCSRF
			cp.csrfLock.Unlock()
		}
	}()

	return cp
}

func (cp *csrfprovider) getCSRF() string {
	cp.csrfLock.RLock()
	resCSRF := cp.csrf
	cp.csrfLock.RUnlock()
	return resCSRF
}

func fetchCSRF() (string, error) {
	req, _ := http.NewRequest(http.MethodGet, "https://instagram.com", nil) //nolint:noctx
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get insta CSRF: %w", err)
	}
	defer resp.Body.Close()

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "csrftoken" {
			return cookie.Value, nil
		}
	}

	return "", ErrNotFound
}
