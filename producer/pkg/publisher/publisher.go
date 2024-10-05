package publisher

import "github.com/nikitades/carassius-bot/shared/task"

type Publisher interface {
	Publish(task task.Task) error
}
