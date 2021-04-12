package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func ProcessGet(pidPathfile string) (process *os.Process, err error) {
	pidBytes, err := ioutil.ReadFile(pidPathfile)

	if err != nil {
		if e, ok := err.(*os.PathError); ok && e.Err.(syscall.Errno) == syscall.ENOENT {
			err = nil
		}

		return
	}

	pid, err := strconv.Atoi(strings.Replace(string(pidBytes), "\n", "", -1))

	if err != nil {
		return nil, fmt.Errorf("parse PID fail. %s", err)
	}

	process, err = os.FindProcess(pid)

	if err != nil {
		err = fmt.Errorf("find process(PID=%v) fail. %s\n", pid, err)
	}

	return
}

func ProcessIsRunned(pidPathfile string) (is bool, err error) {
	proc, err := ProcessGet(pidPathfile)

	if err != nil || proc == nil {
		return
	}

	err = proc.Signal(syscall.Signal(0))

	if err == nil {
		return true, nil
	}

	return false, nil
}

func ProcessSavePid(pidPathfile string) (err error) {
	pid := os.Getpid()

	err = ioutil.WriteFile(pidPathfile, []byte(strconv.Itoa(pid)), 0664)

	return
}
