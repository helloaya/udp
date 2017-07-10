package main

import (
	"log"
	"udp/file"
)



type  TSReader struct {}
func (r *TSReader) Read(id string) (*file.File, error) {
	if "ABCD" != id {
		return nil, file.ERR_RES_NOT_EXIST
	}

	data := make([]byte, file.SIZE_PIECE * 1200) //TODO读取个什么文件?
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