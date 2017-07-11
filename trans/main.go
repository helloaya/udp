package main

import (
	"log"
	"udp/file"
	"os"
	"io/ioutil"
)



type  TSReader struct {}
func (r *TSReader) Read(id string) (*file.File, error) {
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
	r := &TSReader{}
	mgr,err := MakeChansManager (8888, r)
	if nil != err {
		log.Panic (err)
	}
	
	err = mgr.Run()
	if nil != err {
		log.Panic (err)
	}
}
