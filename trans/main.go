package main

import (
	//"fmt"
	"log"
	"net"
	"github.com/golang/protobuf/proto"
	"udp/msg"
)


type Session struct {
	ChanID uint32
	SessionID uint32
	ClientUUID uint32
	Port int32
}


type ChansManager struct {
	conn *net.UDPConn
	sess map[uint32]*Session
}

func MakeChansManager (listenPort int) (*ChansManager, error) {
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: listenPort})
	if nil != err {
		return nil, err
	}
	m := &ChansManager{
		conn : listener,
	}
	m.sess = make(map[uint32]*Session)
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
			log.Println(err)
			continue
		}

		log.Println ("Req = ", req.ClientUUID, "From", remote)
		///根据ClientUUID检查Session是否已经存在
		s,ok := mgr.sess[req.ClientUUID]
		if !ok {
			///不存在,创建Session
			s = &Session {
				ChanID : 18888,
				SessionID : 0,
				ClientUUID : req.ClientUUID,
				Port : 8889,
			}
			mgr.sess[req.ClientUUID] = s
		}

		///响应ReqChanAck
		resp := &msg.ReqChanAck {
			ClientUUID : s.ClientUUID,
			ChanID : s.ClientUUID,
			ChanPort : s.Port,
		}
		out,err := proto.Marshal (resp)
		if nil != err {
			//忽略这个错误
			log.Println(err)
			continue
		}
		if _, err := mgr.conn.Write (out); nil != err {
			//忽略这个错误
			log.Println(err)
			continue
		}
	}
}


func main() {
	mgr,err := MakeChansManager (8888)
	if nil != err {
		log.Panic (err)
	}
	err = mgr.Run()
	if nil != err {
		log.Panic (err)
	}
}


