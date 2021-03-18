package client

// Task in any moment may be canceled. See Canceled()
type Task interface {
	UUID() string
	ParentUUID() string
	// Status gets status of task.
	Status() string
	// StatusSubscribe subscribes on actual status from queue.
	// If task state changed - task will be canceled.
	// No need listen to task.Done() and status simultaniously.
	// Empty status - task not exists.
	// Closed chan - client stopped
	StatusSubscribe() (status <-chan string)
	// Content gets tasks content from task-manager.
	// nil - task canceled.
	// Chan closed with result.
	Content() (content <-chan []byte)
	// StatusSet sets status of task and new content.
	// If <-done == false - task canceled
	StatusSet(status string, content []byte) (done <-chan bool)
	// Remove removes task from queue.
	// Remove should execute on side that created this task.
	// If <-done == false - task canceled
	Remove() (done <-chan bool)
	// Done returns done channel.
	// If it closed - task canceled. See Err().
	Canceled() (canceled <-chan struct{})
	// Err return reason of task cancel.
	Err() (err error)
	// Reject rejects task for to do by another worker.
	// Method should use if task received by queue.TasksSubscribe().
	Reject() (done <-chan struct{})
}
