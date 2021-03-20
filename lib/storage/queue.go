package storage

type Queue interface {
	ID() uint64
	TaskNew(parentUUID string, status string, content []byte) Task
	// TaskGet gets task by uuid.
	// If nil - task not exists
	TaskGet(uuid string) Task

	// TasksGet returns all tasks in queue
	TasksGet() []Task

	// TaskStatusSet sets new status with content and return new task state
	// If return nil - task with uuid not exists
	TaskStatusSet(uuid string, status string, content []byte) Task

	// TaskRemove removes task from queue.
	// If false - task not exists
	TaskRemove(uuid string) (ok bool)
}
