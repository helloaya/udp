package main

import (
	"log"
	"os"
	"io/ioutil"
	"udp/file"
	"udp/core"
)



type  DemoReader struct {}
func (r *DemoReader) Read(id string) (*file.File, error) {
	reader, err  := os.Open(id)
	if nil != err {
		return nil, err
	}
	defer reader.Close()

	data,err := ioutil.ReadAll (reader)
	if nil != err {
		return nil,err
	}

	f := file.MakeFile (id, data)
	return f, nil
}


func main() {
	r := &DemoReader{}
	mgr,err := core.MakeTunnelManager (8888, r)
	if nil != err {
		log.Panic (err)
	}
	
	err = mgr.Run()
	if nil != err {
		log.Panic (err)
	}
}
