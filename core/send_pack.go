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
	// TODO LOG等级
	//	log.Println ("Send DATA", p.Data.Index)
	//}
}

func sendSubcribeAck(tunnel *Tunnel) {
	subAck := &msg.Pack {
		Type : msg.Pack_SUBCRIBE_ACK,
		SubcribeAck : &msg.Pack_SubcribeAck {
			ResouceID : tunnel.Session.getResourceID(),
			SessionID : tunnel.Session.ID,
			IsAccepted : true,
		},
	}
	sendPack (subAck, tunnel.Conn, tunnel.Session.ClientAddr)
}

func sendData(tunnel *Tunnel) {
	if nil != tunnel.Session {
		///寻找下一个待发送的Data包
		if index, payload, ok := tunnel.Session.getNextPack (); ok {
			pack := &msg.Pack {
				Type : msg.Pack_DATA,
				Data :  &msg.Pack_Data {
					SessionID : tunnel.Session.ID,
					Index 	  : index,
					Payload   : payload,
				},
			}
			sendPack (pack, tunnel.Conn, tunnel.Session.ClientAddr)
		}
	}
}

