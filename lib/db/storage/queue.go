package storage

import (
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"sync"
)

type queue struct {
	*dbQueue

	mu        sync.Mutex
	isStarted bool

	db *gorm.DB

	task struct {
		list sync.Map // map[uuid]*task
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

	var dbts []*dbTask

	q.task.list.Range(func(uuid, it interface{}) bool {
		t := it.(*task)

		if t.ParentUUID == "" {
			dbts = append(dbts, t.dbTask)
		}

		return true
	})

	if len(dbts) != 0 {
		err := q.db.CreateInBatches(dbts, 1000).Error

		if err != nil {
			log.Printf("Queue[%s]: Save tasks(%d) fail. %s\n%+v\n", q.Name, len(dbts), err, dbts)
		}
	} else {
		log.Printf("Queue[%s]: NO TASKS\n", q.Name)
	}

	q.task.list = sync.Map{}
}

func (q *queue) taskNew(parentUUID string, status string, content []byte) *TaskInfo {
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
		QueueID:    q.ID,
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

	return &TaskInfo{
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

		tx := q.db.Begin()

		dbt := &dbTask{UUID: uuid}

		err := tx.Where(dbt).Take(dbt).Error

		if err != nil {
			tx.Rollback()

			if err == gorm.ErrRecordNotFound {
				return nil
			}

			panic(fmt.Errorf("get task `%s` from db fail. %s", uuid, err))
		}

		err = tx.Unscoped().Delete(dbt).Error

		if err != nil {
			tx.Rollback()
			panic(fmt.Errorf("delete task `%s` from db fail. %s", uuid, err))
		}

		tx.Commit()

		dbt.ID = 0

		it = &task{
			dbTask: dbt,
		}

		q.task.list.Store(uuid, it)
		break
	}

	return it.(*task)
}

func (q *queue) taskGetInfo(uuid string) (info *TaskInfo) {
	t := q.taskGet(uuid)

	if t != nil {
		t.mu.Lock()
		defer t.mu.Unlock()
		info = &TaskInfo{
			QueueId:    uint64(q.ID),
			UUID:       t.UUID,
			ParentUUID: t.ParentUUID,
			Status:     t.Status,
			Content:    t.Content,
		}
	}

	return
}

func (q *queue) tasksInfoGet() (tasks []*TaskInfo) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.isStarted {
		panic(fmt.Errorf("get all tasks from `%s` queue fail. storage stopped", q.Name))
	}

	dbts := make([]*dbTask, 0)

	tx := q.db.Begin()

	err := tx.Where(&dbTask{QueueID: q.ID}).Find(&dbts).Error

	if err != nil {
		tx.Rollback()
		panic(fmt.Errorf("get queue `%s` all tasks from db fail. %s", q.Name, err))
	}

	err = tx.Unscoped().Where(&dbTask{QueueID: q.ID}).Delete(&dbTask{QueueID: q.ID}).Error

	if err != nil {
		tx.Rollback()
		panic(fmt.Errorf("delete tasks from db queue `%s` fail. %s", q.Name, err))
	}

	tx.Commit()

	for _, dbt := range dbts {
		it, ok := q.task.list.Load(dbt.UUID)

		if !ok {
			it = &task{
				dbTask: dbt,
			}

			q.task.list.Store(dbt.UUID, it)
		}

		it.(*task).ID = 0
	}

	q.task.list.Range(func(uuid, it interface{}) bool {
		t := it.(*task)
		t.mu.Lock()
		defer t.mu.Unlock()
		tasks = append(tasks, &TaskInfo{
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
			err := q.db.Unscoped().Delete(t.dbTask).Error

			if err != nil {
				panic(fmt.Errorf("remove task `%s` from db queue `%s` fail. %s", t.UUID, q.Name, err))
			}
		}

		q.task.list.Delete(t.UUID)

		return true
	}

	return
}
