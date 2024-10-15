package reddit

type ShredditDataType struct {
	Post struct {
		Type string `json:"type"`
	} `json:"post"`
}

type ShredditDataVideo struct {
	PlaybackMP4S struct {
		Permutations []struct {
			Source struct {
				URL        string `json:"url"`
				Dimensions struct {
					Width  int `json:"width"`
					Height int `json:"height"`
				} `json:"dimensions"`
			} `json:"source"`
		} `json:"permutations"`
	} `json:"playbackMp4s"`
}
