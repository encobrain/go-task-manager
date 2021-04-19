package storage

import "github.com/encobrain/go-task-manager/lib/db/storage"

type Task interface {
	// Canceled returns channel of actuality of task state
	// If it closed - task state changed and datas not actual
	Canceled() <-chan struct{}
	UUID() string
	ParentUUID() string
	Status() string
	Content() []byte
}

func NewTask(info *storage.TaskInfo) Task {
	t := &task{
		info:     info,
		canceled: make(chan struct{}),
	}

	return t
}

type task struct {
	info     *storage.TaskInfo
	canceled chan struct{}
}

func (t *task) cancel() {
	defer func() { recover() }()
	close(t.canceled)
}

func (t *task) statusSet(status string, content []byte) *task {
	t.cancel()

	return NewTask(&storage.TaskInfo{
		QueueId:    t.info.QueueId,
		UUID:       t.info.UUID,
		ParentUUID: t.info.ParentUUID,
		Status:     status,
		Content:    content,
	}).(*task)
}

func (t *task) Canceled() <-chan struct{} {
	return t.canceled
}

func (t *task) UUID() string {
	return t.info.UUID
}

func (t *task) ParentUUID() string {
	return t.info.ParentUUID
}

func (t *task) Status() string {
	return t.info.Status
}

func (t *task) Content() []byte {
	return t.info.Content
}
