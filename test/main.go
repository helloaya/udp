package main

import (
	"fmt"
	"net"
	"github.com/golang/protobuf/proto"
	"udp/msg"
)


func SendReqChan () {
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8888}
	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	m := &msg.ReqChan {ClientUUID : 1024}
	out,err := proto.Marshal (m)
	if nil != err {
		fmt.Println (err)
		return
	}
	conn.Write (out)
}

func main() {
	SendReqChan ();
}
