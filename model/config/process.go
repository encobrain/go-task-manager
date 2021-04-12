package config

import "runtime"

type Process struct {
	Name string `yaml:"name"   env:"PROCESS_NAME"   long:"name"   description:"Process name"`
	Run  struct {
		PidPathfile string `yaml:"pid_pathfile"   env:"PROCESS_PID_PATHFILE"   long:"run.pidPathfile"   description:"Pathfile for store process id"`
	} `yaml:"run" namespace:"run"`
	Threads struct {
		Max int `yaml:"max"   env:"PROCESS_THREADS_MAX"   long:"threads.max"   description:"Use max threads. 0=cpu count"   default:"0"`
	} `yaml:"threads"`
}

func (p Process) Pathfile() string {
	return "process.yaml"
}

func (p *Process) Apply() {
	c := p.Threads.Max
	if c == 0 {
		c = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(c)
}
