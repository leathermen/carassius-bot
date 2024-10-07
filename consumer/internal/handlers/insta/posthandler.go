package insta

import "regexp"

func getPostID(url string) (string, bool) {
	pattern := regexp.MustCompile(`(?i)/p/([A-Za-z0-9_-]+)`)
	if match := pattern.FindStringSubmatch(url); len(match) > 1 {
		return match[1], true
	}

	return "", false
}
