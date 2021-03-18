package storage

type Task interface {
	// Canceled returns channel of actuality of task state
	// If it closed - task state changed and datas not actual
	Canceled() <-chan struct{}
	UUID() string
	ParentUUID() string
	Status() string
	Content() []byte
}
