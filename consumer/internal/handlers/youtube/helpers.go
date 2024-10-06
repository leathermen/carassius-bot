package youtube

import (
	"regexp"
)

func extractShortVideoID(link string) (string, bool) {
	// Определение регулярного выражения для поиска идентификатора видео
	regex := regexp.MustCompile(`(?:youtu\.be/|/v/|/u/\w/|/embed/|/shorts/|/e/|/at/|/vi/|watch\?v=|/m/)([^/?&]+)`)

	// Ищем совпадение в ссылке
	matches := regex.FindStringSubmatch(link)
	if len(matches) != 2 {
		return "", false
	}

	// Возвращаем найденный идентификатор видео
	return matches[1], true
}
