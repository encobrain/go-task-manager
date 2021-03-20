package storage

type Queue interface {
	ID() uint64
	TaskNew(parentUUID string, status string, content []byte) Task
	// TaskGet gets task by uuid.
	// If nil - task not exists
	TaskGet(uuid string) Task

	TasksGet() []Task

	// TaskStatusSet sets new status with content and return new task state
	// If return nil - task with uuid not exists
	TaskStatusSet(uuid string, status string, content []byte) Task
}
