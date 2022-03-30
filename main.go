package main

import (
	"fmt"
	"gowatch/core"
	"os"
	"os/exec"
	"os/signal"
	syscall "syscall"
	"time"
)

func init() {
	core.MainInit()
}

func main() {

	signalChan := make(chan os.Signal, 1)
	go func() {
		for {
			select {
			case s := <-signalChan:
				if s == syscall.SIGHUP {
					core.DaemonsRestart(s)
				} else {
					core.DaemonsStop(s)
				}
			}
		}
	}()
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGTERM)

	core.Daemons().Start()
}
func daemonProgram(path string) {
	cmd := exec.Command(path)
	time.AfterFunc(time.Second, func() {
		fmt.Printf("PID %d\n", cmd.Process.Pid)
		fmt.Printf("%+v\n", cmd.ProcessState)
	})
	cmd.Run()
}
