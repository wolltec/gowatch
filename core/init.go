package core

import "github.com/wolltec/golibaray/config"

func MainInit()  {
	config.EnableFlagArgs("./gowatch.yml")
	initDingAlert(&dingAlert{
		env:            config.GetString("dingalert.name"),
		flag:           config.GetString("dingalert.flag"),
		token:          config.GetString("dingalert.token"),
		defaultMobiles: config.GetStrings("dingalert.mobiles"),
	})
}