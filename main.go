package main

import (
	"fmt"
	"gowatch/core"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func init()  {
	core.MainInit()
}

func main()  {

	signalChan := make(chan os.Signal, 1)
	go func() {
		s:=<-signalChan
		log.Println("【终止】接收到信号",s)
		core.DaemonsStop()
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	core.Daemons().Start()
}
func daemonProgram(path string)  {
	cmd := exec.Command(path)
	time.AfterFunc(time.Second, func() {
		fmt.Printf("PID %d\n", cmd.Process.Pid)
		fmt.Printf("%+v\n", cmd.ProcessState)
	})
	cmd.Run()
}
