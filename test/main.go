package main

import (
	"log"
	"net"
	"github.com/golang/protobuf/proto"
	"udp/msg"
	//"time"
)

var (
	RESOURCE_ID = "1496830040.ts"
	RESOURCE_START = 0
	RESOURCE_END = 960
)

func SendReqChan() (uint32, uint32){
	conn, err := net.DialUDP("udp", 
						&net.UDPAddr{IP: net.IPv4zero, Port: 0}, 
						&net.UDPAddr{IP: net.ParseIP("127.0.0.1"), 
						//dstAddr := &net.UDPAddr{IP: net.ParseIP("172.16.0.120"), Port: 8888}
						Port: int(8888)})
	defer conn.Close()

	m := &msg.ReqChan {ClientID : 1024}
	out,err := proto.Marshal (m)
	if nil != err {
		log.Panic(err)
	}
	conn.Write (out)
	buffer := make([]byte, 1500)
	n, err := conn.Read (buffer)
	if nil != err {
		log.Panic(err)
	}

	resp := &msg.ReqChanAck {}
	if err := proto.Unmarshal (buffer[:n], resp); nil != err {
		log.Panic(err)
	}
	return resp.ChanID, resp.ChanPort
}

func SendPack(p* msg.Pack, conn *net.UDPConn) {
	out,err := proto.Marshal (p)
	if nil != err {
		log.Panic(err)
	}
	conn.Write (out)
	log.Println ("Send", *p)
}

func Subcribe(chanID uint32, conn *net.UDPConn) uint32{
	log.Printf("Subcribe chanID=%dn",chanID)

	/// 发送Subcribe
	sub  := &msg.Pack {}
	sub.Type = msg.Pack_SUBCRIBE
	sub.Subcribe = &msg.Pack_Subcribe {}
	sub.Subcribe.ChanID = chanID
	sub.Subcribe.ResouceID = RESOURCE_ID
	sub.Subcribe.Start = RESOURCE_START
	sub.Subcribe.End = RESOURCE_END
	SendPack (sub, conn)

	/// 等待SubscribeAck
	buffer := make([]byte, 1500)
	n, err := conn.Read (buffer)
	if nil != err {
		log.Panic(err)
	}
	subAck := &msg.Pack {}
	if err := proto.Unmarshal (buffer[:n], subAck); nil != err {
		log.Panic(err)
	}
	log.Println ("Recv", subAck)
	return subAck.SubcribeAck.SessionID
}

func RecvData (sessionID uint32, conn *net.UDPConn) {
	for {
		
	}
}


func main() {
	chanID,port := SendReqChan ()
	conn, err := net.DialUDP("udp", 
								&net.UDPAddr{IP: net.IPv4zero, Port: 0}, 
								&net.UDPAddr{IP: net.ParseIP("127.0.0.1"), 
								Port: int(port)})
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()
	sessionID := Subcribe (chanID, conn)
	RecvData (sessionID, conn)
}



