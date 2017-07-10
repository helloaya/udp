package main

import (
	"log"
)



func main() {
	mgr,err := MakeChansManager (8888)
	if nil != err {
		log.Panic (err)
	}
	
	err = mgr.Run()
	if nil != err {
		log.Panic (err)
	}
}