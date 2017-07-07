package main

import (
	"fmt"
	"net"
	"github.com/golang/protobuf/proto"
	"udp/msg"
)


func SendReqChan () {
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	//dstAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8888}
	dstAddr := &net.UDPAddr{IP: net.ParseIP("172.16.0.120"), Port: 8888}
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
	buffer := make([]byte, 1500)
	n, err := conn.Read (buffer)
	if nil != err {
		resp := &msg.ReqChanAck {}
		if err := proto.Unmarshal (buffer[:n], resp); nil != err {
			//忽略这个错误
			fmt.Println (err)
			return
		}
		fmt.Println (resp)
	}

}

func main() {
	SendReqChan ();
}
