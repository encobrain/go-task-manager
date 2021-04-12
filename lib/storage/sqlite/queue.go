package sqlite

import (
	"fmt"
	"github.com/encobrain/go-task-manager/lib/storage"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"sync"
)

type queue struct {
	*dbQueue

	mu        sync.Mutex
	isStarted bool

	db *gorm.DB

	task struct {
		list sync.Map // map[uuid]*task
		all  bool
	}
}

func (q *queue) start() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.isStarted = true
}

func (q *queue) stop() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.isStarted = false
}

func (q *queue) taskNew(parentUUID string, status string, content []byte) *storage.TaskInfo {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.isStarted {
		panic(fmt.Errorf("create new task fail. storage stopped"))
	}

	uuid, err := uuid.NewUUID()

	if err != nil {
		panic(fmt.Errorf("generate uuid fail. %s", err))
	}

	dbt := &dbTask{
		UUID:       uuid.String(),
		ParentUUID: parentUUID,
		Status:     status,
		Content:    content,
	}

	t := &task{
		dbTask:  dbt,
		updated: true,
	}

	q.task.list.Store(dbt.UUID, t)

	return &storage.TaskInfo{
		QueueId:    uint64(q.ID),
		UUID:       dbt.UUID,
		ParentUUID: dbt.ParentUUID,
		Status:     dbt.Status,
		Content:    dbt.Content,
	}
}

func (q *queue) taskGet(uuid string) *task {
	it, _ := q.task.list.Load(uuid)

	for it == nil {
		q.mu.Lock()
		defer q.mu.Unlock()

		if !q.isStarted {
			panic(fmt.Errorf("get task `%s` fail. storage stopped", uuid))
		}

		it, _ = q.task.list.Load(uuid)

		if it != nil {
			break
		}

		if q.task.all {
			return nil
		}

		dbt := &dbTask{UUID: uuid}

		err := q.db.Where(dbt).Take(dbt).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil
			}

			panic(fmt.Errorf("get task `%s` from db fail. %s", uuid, err))
		}

		it = &task{
			dbTask: dbt,
		}

		q.task.list.Store(uuid, it)
		break
	}

	return it.(*task)
}

func (q *queue) taskGetInfo(uuid string) (info *storage.TaskInfo) {
	t := q.taskGet(uuid)

	if t != nil {
		t.mu.Lock()
		defer t.mu.Unlock()
		info = &storage.TaskInfo{
			QueueId:    uint64(q.ID),
			UUID:       t.UUID,
			ParentUUID: t.ParentUUID,
			Status:     t.Status,
			Content:    t.Content,
		}
	}

	return
}

func (q *queue) tasksInfoGet() (tasks []*storage.TaskInfo) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.isStarted {
		panic(fmt.Errorf("get all tasks from `%s` queue fail. storage stopped", q.Name))
	}

	if !q.task.all {
		dbts := make([]*dbTask, 0)

		err := q.db.Where(&dbTask{QueueID: q.ID}).Find(&dbts).Error

		if err != nil {
			panic(fmt.Errorf("get queue `%s` all tasks from db fail. %s", q.Name, err))
		}

		for _, dbt := range dbts {
			if _, ok := q.task.list.Load(dbt.UUID); !ok {
				it := &task{
					dbTask: dbt,
				}

				q.task.list.Store(dbt.UUID, it)
			}
		}

		q.task.all = true
	}

	q.task.list.Range(func(uuid, it interface{}) bool {
		t := it.(*task)
		t.mu.Lock()
		defer t.mu.Unlock()
		tasks = append(tasks, &storage.TaskInfo{
			QueueId:    uint64(q.ID),
			UUID:       t.UUID,
			ParentUUID: t.ParentUUID,
			Status:     t.Status,
			Content:    t.Content,
		})

		return true
	})

	return
}

func (q *queue) taskStatusSet(uuid string, status string, content []byte) (ok bool) {
	t := q.taskGet(uuid)

	if t != nil {
		t.statusSet(status, content)
		return true
	}

	return
}

func (q *queue) taskRemove(uuid string) (ok bool) {
	t := q.taskGet(uuid)

	if t != nil {
		q.mu.Lock()
		defer q.mu.Unlock()

		if !q.isStarted {
			panic(fmt.Errorf("remove task `%s` from queue `%s` fail. storage stopped", t.UUID, q.Name))
		}

		if t.ID != 0 {
			err := q.db.Delete(t.dbTask).Error

			if err != nil {
				panic(fmt.Errorf("remove task `%s` from db queue `%s` fail. %s", t.UUID, q.Name, err))
			}
		}

		q.task.list.Delete(t.UUID)

		return true
	}

	return
}
