package core

import (
	"bytes"
	"github.com/wolltec/golibaray/config"
	"log"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

type DaemonConfig struct {
	Name string
	Program string
	Retry int
}
var Programs []*exec.Cmd
var Stop = false

type DaemonConfigs []DaemonConfig

func Daemons(key ...string) *DaemonConfigs {
	log.Println("加载配置")
	if len(key) == 0 {
		key = append(key, "daemons")
	}
	daemons := new(DaemonConfigs)
	err := config.GetStruct(key[0], daemons)
	if err != nil{
		log.Printf("配置加载失败 %e \n", err)
	}
	return daemons
}

func (d *DaemonConfigs) Start()  {
	var wg sync.WaitGroup
	if len(*d) < 1{
		log.Println("未读取到配置内容")
		return
	}

	log.Printf("启动守护服务【%d】项...\n",len(*d))
	for _,p := range *d{
		if p.Retry > 0{
			p.Retry++
		}
		wg.Add(1)

		go p.daemon(&wg)
	}
	wg.Wait()
}
func DaemonsStop()  {
	Stop = true
	for _,cmd := range Programs{
		kill(cmd)
	}
	log.Println("父进程退出")
}

//守护逻辑
func (p DaemonConfig) daemon(wg *sync.WaitGroup)  {
	defer wg.Done()
	programConfig := strings.Split(p.Program," ")
	program := programConfig[0]
	args := programConfig[1:]
	retry := 1
	restart := "启动"
	for p.Retry != 0 && !Stop{
		log.Printf("【%s】[%d] %s\n", restart, retry, p.Name)
		notice := listen(program, args...)
		log.Printf("【异常】%s | 捕获到程序异常退出\n------------------ 异常信息 ------------------\n%s\n--------------------------------------------\n", p.Name, notice)
		//发送告警
		//if DingAlert.SendPanic(notice) == nil{
		//	log.Printf("【告警】%s | 已推送异常告警\n", p.Name)
		//}
		p.Retry-- //重试计数
		retry++
		restart = "重试"
		time.Sleep(time.Second) //控制重启频率
	}
	log.Printf("【进程结束】%s | 已执行最大次数重启\n", p.Name)
}

func kill(cmd *exec.Cmd) {
	err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			if err == nil{
				log.Printf("子进程 %d 退出\n", cmd.Process.Pid)
			}else{

				log.Printf("子进程 %d 退出 异常\n", cmd.Process.Pid, err)
			}
}

// 监听进程
func listen(program string, args ...string) (notice string) {
	daemon := exec.Command(program, args...)
	daemon.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	var panic bytes.Buffer
	daemon.Stderr = &panic
	go func() {
		for	true {
			time.Sleep(time.Millisecond * 500)
			if daemon.Process == nil{
				log.Println("等待子进程PID")
				continue
			}
			log.Printf("【%d】子进程启动成功", daemon.Process.Pid )
			Programs = append(Programs, daemon)
			return
		}
	}()
	err := daemon.Run()
	if err != nil{
		notice = err.Error()+"\n"
	}
	notice += panic.String()
	return
}