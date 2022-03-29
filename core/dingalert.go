package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var host = "https://oapi.dingtalk.com/robot/send?access_token="
var Token string
var AtMobiles []string

var DingAlert *dingAlert

type dingAlert struct {
	env            string
	token          string
	flag           string
	defaultMobiles []string
}
type ding struct {
	Msgtype string   `json:"msgtype,omitempty"`
	At      dingAt   `json:"at"`
	Text    dingText `json:"text"`
}
type dingText struct {
	Content string `json:"content,omitempty"`
}
type dingAt struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
}
type dingResponse struct {
	Errcode int    `json:"errcode,omitempty"`
	Errmsg  string `json:"errmsg,omitempty"`
}

func initDingAlert(conf *dingAlert)  {
	DingAlert = conf
}
func (l *dingAlert) Send(title string, content string, mobiles ...string) error {
	if content == "" {
		return errors.New("empty content")
	}
	data := ding{
		Msgtype: "text",
		Text: dingText{
			Content: fmt.Sprintf("%s %s # %s\n%s",
				l.flag, title, l.env, content),
		},
	}
	if len(mobiles) > 0 {
		data.At.AtMobiles = mobiles
	} else if len(l.defaultMobiles) > 0 {
		data.At.AtMobiles = l.defaultMobiles
	}
	result, err := post(host+l.token, data)
	if err != nil {
		return err
	}
	if result.Errcode > 0 {
		return errors.New(result.Errmsg)
	}
	return nil
}
func (l *dingAlert) SendPanic(content string, mobiles ...string) error {
	now := "时间："+time.Now().Format("2006-02-01 15:04:05")+"\n详细信息：\n"
	return l.Send("进程中断 | 重要", now+content, mobiles...)
}
func post(url string, data interface{}) (*dingResponse, error) {
	jsonData, _ := json.Marshal(data)
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := new(dingResponse)
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	if result.Errcode !=0 {
		fmt.Println("Error", result)
	}
	return result, nil
}
