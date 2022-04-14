package core

import (
	"bytes"
	"fmt"
	"github.com/wolltec/golibaray/config"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

type DaemonConfig struct {
	Name         string
	Program      string
	RetryBalance int
	MaxRetry     int `yaml:"max_retry"`
	Retry        int
}
type DaemonConfigs []DaemonConfig

// programs 子进程管理
var programs []*exec.Cmd

// 强制退出标记
var stop, restart = false, false

//监控&告警 频率
var sendRate = time.Minute
var sendLock = false

func Daemons(key ...string) *DaemonConfigs {
	log.Println("加载配置")
	if len(key) == 0 {
		key = append(key, "daemons")
	}
	daemons := new(DaemonConfigs)
	err := config.GetStruct(key[0], daemons)
	if err != nil {
		log.Printf("配置加载失败 %e \n", err)
	}
	return daemons
}

func (d *DaemonConfigs) Start() {
	var wg sync.WaitGroup
	if len(*d) < 1 {
		log.Println("未读取到配置内容")
		return
	}

	log.Printf("启动服务 任务数 %d\n", len(*d))
	for i, p := range *d {
		if p.MaxRetry > 0 {
			p.MaxRetry++
		}
		p.RetryBalance = p.MaxRetry
		wg.Add(1)

		go p.daemon(&wg, i+1)
	}
	wg.Wait()
}
func DaemonsStop(s os.Signal) {
	log.Println("【中止】接收到信号", s)
	stop = true
	for _, cmd := range programs {
		kill(cmd)
	}
}

func DaemonsRestart(s os.Signal) {
	log.Println("【重启】接收到信号", s)
	restart = true
	for _, cmd := range programs {
		kill(cmd)
	}
}

//守护逻辑
func (p DaemonConfig) daemon(wg *sync.WaitGroup, number int) {
	defer wg.Done()
	programConfig := strings.Split(p.Program, " ")
	program := programConfig[0]
	args := programConfig[1:]
	actionName := "启动"
	for p.RetryBalance != 0 {
		var info string
		if p.Retry == 0 {
			info = fmt.Sprintf("【%s】任务 %d - %s", actionName, number, p.Name)
		} else {
			info = fmt.Sprintf("【%s】任务 %d - %s | 重试 %d", actionName, number, p.Name, p.Retry)
		}
		notice := listen(info, program, args...)
		if stop { //强制退出
			log.Printf("【结束】父进程退出")
			return
		}
		if restart { //重启，清零，不告警
			p.Retry = 0
			p.RetryBalance = p.MaxRetry
			programs = make([]*exec.Cmd, 0)
			restart = false
			continue
		}
		log.Printf("【异常】%s | 捕获到程序异常退出\n------------------ 异常信息 ------------------\n%s\n--------------------------------------------\n", p.Name, notice)
		//发送告警
		go sendDingPanic(p.Name, notice)
		p.RetryBalance-- //重试计数
		p.Retry++
		actionName = "重试"
		time.Sleep(time.Second)
	}
	log.Printf("【进程结束】%s | 已执行最大次数重启\n", p.Name)
}

//退出机制
func kill(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	err := syscall.Kill(cmd.Process.Pid, syscall.SIGKILL)
	if err == nil {
		log.Printf("【中止】子进程 %d 退出\n", cmd.Process.Pid)
	} else {
		log.Printf("【中止】子进程 %d 退出 异常 %e\n", cmd.Process.Pid, err)
	}
}

// 监听进程
func listen(info string, program string, args ...string) (notice string) {
	daemon := exec.Command(program, args...)
	//daemon.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var stderr bytes.Buffer //捕捉标准错误
	daemon.Stderr = &stderr
	output, _ := os.Create(config.GetString("logs.output"))
	daemon.Stdout = output
	defer output.Close()
	go func() {
		for true {
			time.Sleep(time.Millisecond * 500)
			if daemon.Process == nil {
				//log.Println("等待子进程PID...")
				continue
			}
			log.Printf("%s | PID %d", info, daemon.Process.Pid)
			programs = append(programs, daemon)
			return
		}
	}()
	err := daemon.Run()
	if err != nil {
		notice = err.Error() + "\n"
	}
	notice += stderr.String()
	return
}

func sendDingPanic(name, content string) {
	if sendLock {
		log.Printf("【告警】%s | 推送过于频繁，已跳过\n", name)
		return
	}
	if err := DingAlert.SendPanic(name, content); err != nil {
		log.Println("告警异常", err)
	}
	sendLock = true
	log.Printf("【告警】%s | 已推送异常告警\n", name)

	time.AfterFunc(sendRate, func() {
		sendLock = false
	})
}
