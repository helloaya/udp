package main

import (
	"fmt"
	"time"
	"net"
	"github.com/golang/protobuf/proto"
	"udp/msg"
)


var TOTAL_PACKS uint32 = 128 * 10 
func runServer (port int) {
	fmt.Println ("runServer", port)
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: port})
	if err != nil {
		fmt.Println(err)
		return
	}

	err = listener.SetReadDeadline(time.Now ().Add (time.Second * 10))
	if nil != err {
		fmt.Println (err)
		return
	}

	data := make([]byte, 2048)
	var total uint32 = 0
	for {
		n, _, err := listener.ReadFromUDP(data)
			if err != nil {
			//type assertion
			nerr, ok := err.(net.Error)
			if ok && nerr.Timeout () {
				fmt.Println ("Timeout")
			} else {
				fmt.Println (err.Error())
			}
			return
		}
		m := &msg.Pack {
		}
		err = proto.Unmarshal (data[:n], m)
		if nil != err {
			fmt.Println(err)
			return
		}
		/*
		if m.Index == TOTAL_PACKS {
			break
		}*/
		total += 1
	}
	fmt.Println ("Loss", float32(TOTAL_PACKS - total)/float32(TOTAL_PACKS))
}

/*
func runClient () {
	fmt.Println ("runClient")
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
	var n uint32 = TOTAL_PACKS
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
	m.Index = n
	i = 0
	for i<10 {
		out,err := proto.Marshal (m)
		if nil != err {
			fmt.Println (err)
		}
		conn.Write (out)
		i += 1
		time.Sleep (time.Second)
	}
}
*/
