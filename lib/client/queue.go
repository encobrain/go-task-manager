package client

type Queue interface {
	// TaskNew creates new task with parrentUUID. parentUUID may be empty.
	// Chan closed with result.
	// If nil - client stopped.
	TaskNew(parentUUID string, status string, content []byte) (task <-chan Task)
	// TaskGet gets task with uuid.
	// Chan closed with result.
	// If nil - task not exists.
	TaskGet(uuid string) (task <-chan Task)
	// TasksSubscribe subscribes to get tasks from queue for process.
	// If paretnUUID not empty - subscribe onlyon tasks with that parent uuid
	// If nil - client stopped.
	TasksSubscribe(parentUUID string) (tasks <-chan Task)
	// TasksGet gets all tasks with/or without parentUUID in queue.
	// If nil - client stopped.
	// Chan closed with result.
	TasksGet(parentUUID string) (tasks <-chan []Task)
}
