package storage

import "sync"

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

func NewQueue(stor Storage, info *QueueInfo) Queue {
	q := &queue{
		storage: stor,
		info:    info,
	}

	return q
}

type queue struct {
	info    *QueueInfo
	storage Storage

	mu   sync.Mutex
	task map[string]*task
}

func (q *queue) ID() uint64 {
	return q.info.Id
}

func (q *queue) TaskNew(parentUUID string, status string, content []byte) Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	ti := q.storage.TaskNew(q.info.Id, parentUUID, status, content)

	t := NewTask(ti)

	q.task[ti.UUID] = t.(*task)

	return t
}

func (q *queue) taskGet(uuid string) Task {
	t := q.task[uuid]

	if t == nil {
		ti := q.storage.TaskGet(q.info.Id, uuid)

		if ti == nil {
			return nil
		}

		t = NewTask(ti).(*task)

		q.task[ti.UUID] = t
	}

	return t
}

func (q *queue) TaskGet(uuid string) Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.taskGet(uuid)
}

func (q *queue) TasksGet() (tasks []Task) {
	q.mu.Lock()
	defer q.mu.Unlock()

	tis := q.storage.TasksGet(q.info.Id)

	for _, ti := range tis {
		t := q.task[ti.UUID]

		if t == nil {
			t = NewTask(ti).(*task)
			q.task[ti.UUID] = t
		}

		tasks = append(tasks, t)
	}

	return
}

func (q *queue) TaskStatusSet(uuid string, status string, content []byte) Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	it := q.taskGet(uuid)

	if it == nil {
		return nil
	}

	if !q.storage.TaskStatusSet(q.info.Id, uuid, status, content) {
		return nil
	}

	t := it.(*task)

	t = t.statusSet(status, content)

	q.task[uuid] = t

	return t
}

func (q *queue) TaskRemove(uuid string) (ok bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	it := q.taskGet(uuid)

	if it == nil {
		return
	}

	if !q.storage.TaskRemove(q.info.Id, uuid) {
		return
	}

	t := it.(*task)
	t.cancel()

	delete(q.task, uuid)

	return true
}
