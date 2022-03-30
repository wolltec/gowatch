### 项目代码
GIT仓库地址：
[https://git.addnewer.com/middleware/rmlibs/tree/master/gowatch](https://git.addnewer.com/middleware/rmlibs/tree/master/gowatch)
### 配置说明
```yaml
daemons:
  -
  # 被守护程序名称
  name: "GoWatch"
  # 被守护程序的启动命令
  program: "./gowatch-linux -c local.yml"
  # 守护期间最大重启次数，小于0代表无限次重启
  max_retry: -1
  #	当程序意外中断，守护服务会尝试重新启动程序，每秒最多执行一次重启


dingalert:
  #告警标记，取自钉钉机器人
  flag: "# "
  # 告警群 Token
  token: "998e143de293624e08bc998ae4e94b3e0f918ca00e701042d1f196b57ea3fcc4"
  # 需要重点@同学的钉钉手机号，批量
  mobiles:
    - "18064035445"
```
### 启动方式
linux环境下，通过nohup启动gowatch后台进程，默认加载 ./gowatch.yml 配置文件，输出日志到 gowatch.log
```shell
nohup gowatch路径 >> gowatch.log 2>&1 &
```
### CI/CD配置
> 暂不支持被守护程序的热重启
> 项目首次配置时，通过手动执行上述启动方式来启动gowatch，后续项目代码有更新时，可以通过执行
> pkill -1 gowatch
> gowatch监听 SIGHUB/1 信号，收到信号后会执行重启

#### shell 示例
为了方便同一台机器的多个gowatch监控，这里将gowatch更名为 gowatch-adoctopus-rps 便于定位
```shell
#!/bin/bash
#查找 gowatch-adoctopus-rps 的 pid
pid=$(ps aux|grep gowatch-adoctopus-rps |grep -v grep|awk '{print $2}')
if [ "$pid" = "" ] #如果找不到
then
  #启动 gowatch-adoctopus-rps
  nohup ./daemon/gowatch-adoctopus-rps -c ./daemon/gowatch.yml >> /data/modules/adoctopus_rps_sandbox/gowatch.log 2>&1 &
else
  #如果找到了pid（gowatch-adoctopus-rps运行中）,执行 kill -1 pid
  ps aux|grep gowatch-adoctopus-rps |grep -v grep|awk '{print $2}'|xargs kill -1
fi
```
