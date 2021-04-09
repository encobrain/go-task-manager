package storage

type QueueInfo struct {
	Id   uint64
	Name string
}

type TaskInfo struct {
	QueueId    uint64
	UUID       string
	ParentUUID string
	Status     string
	Content    []byte
}

type Storage interface {
	// QueueGetOrCreate gets queue by name or creates it if not exists
	QueueGetOrCreate(name string) *QueueInfo
	// QueueGet gets queue by id.
	// If nil - queue not exists
	QueueGet(id uint64) *QueueInfo
	// TaskNew creates new task in queue
	// If nil - queueId invalid
	TaskNew(queueId uint64, parentUUID string, status string, content []byte) *TaskInfo
	// TaskGet gets task by uuid.
	// If nil - task not exists or queueId invalid
	TaskGet(queueId uint64, uuid string) *TaskInfo
	// TasksGet returns all tasks in queue
	TasksGet(queueId uint64) []*TaskInfo
	// TaskStatusSet sets new status with content.
	// If ok==fale - task with uuid not exists or queueId invalid
	TaskStatusSet(queueId uint64, uuid string, status string, content []byte) (ok bool)
	// TaskRemove removes task from queue.
	// If ok==false - task not exists or queueId invalid
	TaskRemove(queueId uint64, uuid string) (ok bool)
}
