package storage

import "gorm.io/gorm"

type dbQueue struct {
	gorm.Model

	Name string
}

func (dbQueue) TableName() string {
	return "queues"
}

type dbTask struct {
	gorm.Model

	QueueID    uint
	UUID       string
	ParentUUID string
	Status     string
	Content    []byte
}

func (dbTask) TableName() string {
	return "tasks"
}
