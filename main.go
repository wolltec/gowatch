package main

import "gowatch/core"

func main()  {

	//for range time.Tick(time.Second){
	//	fmt.Println(time.Now().String())
	//}

	core.Daemons().Start()
}
