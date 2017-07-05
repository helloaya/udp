package main

import (
	"fmt"
	"time"
	"net"
	"flag"
	"github.com/golang/protobuf/proto"
	"udp/msg"
)

func runServer () {
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 9999})
	if err != nil {
		fmt.Println(err)
		return
	}
	data := make([]byte, 2048)
	for {
		_, _, err := listener.ReadFromUDP(data)
		if err != nil {
			fmt.Printf("error during read: %s", err)
		}
		m := &msg.Pack {}
		err = proto.Unmarshal (data, m)
		if nil != err {
			fmt.Println(err)
			return
		}
		fmt.Println (m)
	}
}

func runClient () {
	m := &msg.Pack {
		SessionID : 1,
		Index : 0,
		Data : make([]byte, 1400),
	}

	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: net.ParseIP("172.16.0.120"), Port: 9999}
	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	var i uint32 = 0
	var n uint32 = 1000
	s := time.Now ()
	w := uint32(time.Second) / n
	fmt.Println (time.Duration(w))
	for i < n {
		m.Index = i
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

func main() {
	runType := flag.String ("type", "server", "running type")
	flag.Parse()

	if "server" == *runType {
		runServer ()
	}

	if "client" == *runType {
		runClient ()
	}
}
