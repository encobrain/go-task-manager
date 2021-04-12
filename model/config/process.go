package config

type Process struct {
	Name string `yaml:"name"   env:"PROCESS_NAME"   long:"process.name"   description:"Process name"`
	Run  struct {
		PidPathfile string `yaml:"pid_pathfile"   env:"PROCESS_PID_PATHFILE"   long:"pidPathfile"   description:"Pathfile for store process id"`
	} `yaml:"run" group:"Run options" namespace:"run"`
	Threads struct {
		Max int `yaml:"max"   env:"PROCESS_THREADS_MAX"   long:"process.threads.max"   description:"Use max threads. 0=cpu count"   default:"0"`
	} `yaml:"threads" group:"Threads options" namespace:"threads"`
}

func (p Process) Pathfile() string {
	return "process.yaml"
}
