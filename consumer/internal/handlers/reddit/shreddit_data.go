package reddit

type ShredditData struct {
	Post struct {
		Type string `json:"type"`
	} `json:"post"`
}
