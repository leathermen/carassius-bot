package task

type MediaType uint8

const (
	Twitter MediaType = iota
	Insta   MediaType = iota
	Reddit  MediaType = iota
)

type Task struct {
	MediaType MediaType
	Url       string
}
