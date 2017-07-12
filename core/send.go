package core


import (
	"log"
	"net"
	"udp/msg"
	"github.com/golang/protobuf/proto"
)



func sendReqTunnelAck(tunnel *Tunnel, conn *net.UDPConn, remote *net.UDPAddr) {
	resp := &msg.ReqTunnelAck {
			ClientID : tunnel.ClientID,
			TunnelID : tunnel.TunnelID,
			TunnelPort : uint32(tunnel.Conn.LocalAddr().(*net.UDPAddr).Port),
		}
	out,err := proto.Marshal (resp)
	if nil != err {
		log.Panic(err)
	}
	if _, err := conn.WriteTo (out, remote); nil != err {
		log.Panic(err)
	}
}


func sendPack(p* msg.Pack, conn *net.UDPConn, remote *net.UDPAddr) {
	out,err := proto.Marshal (p)
	if nil != err {
		log.Panic(err)
	}
	conn.WriteTo (out, remote)
	if p.Type != msg.Pack_DATA {
		log.Println ("Send", *p)
	}
	// else {
	//	log.Println ("Send DATA", p.Data.Index)
	//}
}


func sendSubcribeAck(tunnel *Tunnel, resourceID string) {
	subAck := &msg.Pack {
		Type : msg.Pack_SUBCRIBE_ACK,
		SubcribeAck : &msg.Pack_SubcribeAck {
			ResouceID : resourceID,
			SessionID : tunnel.SessionID,
			IsAccepted : true,
		},
	}
	sendPack (subAck, tunnel.Conn, tunnel.Remote)
}
