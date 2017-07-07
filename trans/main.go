package main

import (
	"fmt"
	"net"
	"github.com/golang/protobuf/proto"
	"udp/msg"
)




func HandleReqChan(req *msg.ReqChan, remote *net.UDPAddr) error {
	fmt.Println ("Req = ", req.ClientUUID, "From", remote)
	return nil
}


func RunReqChanListener(port int) error {
	fmt.Println ("Start RunReqChanListener", port)
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: port})
	if err != nil {
		return err
	}

	buffer := make([]byte, 128)
	for {
		n, remote, err := listener.ReadFromUDP(buffer)
		if err != nil {
			return err
		}
		req := &msg.ReqChan {}
		if err := proto.Unmarshal (buffer[:n], req); nil != err {
			//忽略这个错误
			fmt.Println(err)
			continue
		}
		if err := HandleReqChan (req, remote); nil != err {
			//忽略这个错误
			fmt.Println(err)
			continue
		}
	}
}


func main() {
	if err := RunReqChanListener (8888); nil != err {
		fmt.Println (err)
	}
}
