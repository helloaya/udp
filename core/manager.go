package core

import (
	"log"
	"net"
	"sync"
	"udp/msg"
	"udp/file"
	"math/rand"
	"github.com/golang/protobuf/proto"
)



type TunnelManager struct {
	listener *net.UDPConn
	tunnels map[uint32]*Tunnel
	chansLock sync.Mutex
	reader file.Reader 
}


func MakeTunnelManager(listenPort int, reader file.Reader) (*TunnelManager, error) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: listenPort})
	if nil != err {
		return nil, err
	}
	m := &TunnelManager{
		listener : conn,
	}
	m.tunnels = make(map[uint32]*Tunnel)
	m.reader = reader
	return m, nil
}


func (mgr *TunnelManager) Run() error {
	buffer := make([]byte, 1500)
	for {
		n, remote, err := mgr.listener.ReadFromUDP(buffer)
		if err != nil {
			return err
		}
		req := &msg.ReqTunnel {}
		if err := proto.Unmarshal (buffer[:n], req); nil != err {
			//忽略这个错误
			log.Println(err)
			continue
		}

		log.Println ("Req=", req.ClientID, "From", remote)
		tunnel, err := mgr.add (req.ClientID)
		if nil != err {
			log.Panic (err)
		}
		///响应ReqChanAck
		sendReqTunnelAck (tunnel, mgr.listener, remote)
		go tunnel.Run (mgr)
	}
}


func (mgr *TunnelManager) add(clientID uint32) (*Tunnel, error){
	mgr.chansLock.Lock ()
	defer mgr.chansLock.Unlock ()
	tunnel,ok := mgr.tunnels[clientID]
	if !ok {
		/// 找不到clientID对应的Tunnel,重新创建一个
		conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
		if nil != err {
			return nil, err
		}
		tunnel = &Tunnel {
			TunnelID : rand.Uint32(), ///TODO 随机数,可能会碰撞
			ClientID : clientID,
			SendIndex : 0,
			Conn : conn,
			Rate : RATE_INIT,
		}
		mgr.tunnels[clientID] = tunnel
		log.Println ("Add Tunnel,Port="," ClientID=", clientID)
	}
	return tunnel,nil
}


func (mgr *TunnelManager) delete(tunnel *Tunnel) {
	mgr.chansLock.Lock ()
	defer mgr.chansLock.Unlock ()
	delete (mgr.tunnels, tunnel.ClientID)
	log.Println ("Delete Tunnel,ClientID=", tunnel.ClientID)
}
