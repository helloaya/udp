package main

import (
	//"fmt"
	"log"
	"net"
	"github.com/golang/protobuf/proto"
	"udp/msg"
)




func HandleReqChan(req *msg.ReqChan, remote *net.UDPAddr) error {
	log.Println ("Req = ", req.ClientUUID, "From", remote)
	return nil
}

type ChansManager struct {
	conn *net.UDPConn
}

func MakeChansManager (listenPort int) (*ChansManager, error) {
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: listenPort})
	if nil != err {
		return nil, err
	}
	m := &ChansManager{
		conn : listener,
	}
	return m, nil
}

func (mgr *ChansManager)Run() error {
	buffer := make([]byte, 1500)
	for {
		n, remote, err := mgr.conn.ReadFromUDP(buffer)
		if err != nil {
			return err
		}
		req := &msg.ReqChan {}
		if err := proto.Unmarshal (buffer[:n], req); nil != err {
			//忽略这个错误
			log.Fatal(err)
			continue
		}
		if err := HandleReqChan (req, remote); nil != err {
			//忽略这个错误
			log.Fatal(err)
			continue
		}
	}
}


func main() {
	mgr,err := MakeChansManager (8888)
	if nil != err {
		log.Fatal (err)
		return
	}
	err = mgr.Run()
	if nil != err {
		log.Fatal (err)
		return
	}
}


