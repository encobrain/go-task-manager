package service

type Status string

const (
	StatusStopped  Status = "stopped"
	StatusStarting        = "starting"
	StatusStarted         = "started"
	StatusStopping        = "stopping"
)

type Service interface {
	// Start starts the service
	Start()
	// Stop stops the service
	Stop()
	// Status gets status of the service
	Status() (status Status)
	// StatusWait gets channel for wait status of service. If it closed - status set
	StatusWait(status Status) (waiter <-chan struct{})
}
