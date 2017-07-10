package main

import (
	"log"
	"net"
	"sync"
	"github.com/golang/protobuf/proto"
	"udp/msg"
	"math/rand"
)



type ChansManager struct {
	listener *net.UDPConn
	chans map[uint32]*Chan
	chansLock sync.Mutex
}



func MakeChansManager(listenPort int) (*ChansManager, error) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: listenPort})
	if nil != err {
		return nil, err
	}
	m := &ChansManager{
		listener : conn,
	}
	m.chans = make(map[uint32]*Chan)
	return m, nil
}


func (mgr *ChansManager) clearChan(clientID uint32) {
	mgr.chansLock.Lock ()
	defer mgr.chansLock.Unlock ()
	delete (mgr.chans, clientID)
}

func (mgr *ChansManager) makeChan(clientID uint32) (*Chan, error){
	mgr.chansLock.Lock ()
	defer mgr.chansLock.Unlock ()
	c,ok := mgr.chans[clientID]
	if !ok {
		c = &Chan {}
		c.ChanID = rand.Uint32() ///TODO 随机数,可能会碰撞
		c.ClientID = clientID
		for {
			c.Port = uint16(rand.Uint32() % 10000 + 10000) ///端口范围 10000 - 19999
			conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: int(c.Port)})
			if nil != err {
				continue
			}
			log.Println ("Make Chan,Port=", c.Port)
			c.Conn = conn
			break
		}
		mgr.chans[clientID] = c
	}
	return c,nil
}

func (mgr *ChansManager)Run() error {
	buffer := make([]byte, 1500)
	for {
		n, remote, err := mgr.listener.ReadFromUDP(buffer)
		if err != nil {
			return err
		}
		req := &msg.ReqChan {}
		if err := proto.Unmarshal (buffer[:n], req); nil != err {
			//忽略这个错误
			log.Println(err)
			continue
		}

		log.Println ("Req=", req.ClientID, "From", remote)
		c, err := mgr.makeChan (req.ClientID)
		if nil != err {
			log.Panic (err)
		}
		///响应ReqChanAck
		resp := &msg.ReqChanAck {
			ClientID : c.ClientID,
			ChanID : c.ChanID,
			ChanPort : uint32(c.Port),
		}
		out,err := proto.Marshal (resp)
		if nil != err {
			//忽略这个错误
			log.Println(err)
			continue
		}
		if _, err := mgr.listener.WriteTo (out, remote); nil != err {
			//忽略这个错误
			log.Println(err)
			continue
		}
		go c.Run (mgr)
	}
}

