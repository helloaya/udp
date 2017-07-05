package main

import (
	"fmt"
	"time"
	"net"
	"github.com/golang/protobuf/proto"
	"udp/msg"
)

func main() {
	m := &msg.Command {
		Type : msg.Command_SUBCRIBE,
		Subcribe : &msg.Command_Subcribe {
			ClientUUID : 1,
			ResouceID : "abcd",
			Start : 0,
			End : 100,
		},
	}

	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: net.ParseIP("172.16.0.120"), Port: 9999}
	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	i := 0
	s := time.Now ()
	n := 1000 * 100
	w := int(time.Second) / n
	fmt.Println (time.Duration(w))
	for i < n {
		out,err := proto.Marshal (m)
		if nil != err {
			fmt.Println (err)
		}
		conn.Write (out)
		i += 1
		time.Sleep (time.Duration(w))
	}
	d := time.Since (s)
	fmt.Println (d)
}
