package client

import "github.com/encobrain/go-task-manager/model/config"

type Client interface {
	// GetQueue get queue. If nil - client stopped.
	// Chan closed with result.
	GetQueue(name string) (queue <-chan Queue)
}

func NewClient(config *config.Client) Client {

}
