package main

import (
	"log"
	"net"
	"github.com/golang/protobuf/proto"
	"udp/msg"
	"time"
)


func SendReqChan() (uint32, uint32){
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8888}
	//dstAddr := &net.UDPAddr{IP: net.ParseIP("172.16.0.120"), Port: 8888}
	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		log.Panic(err)
	}
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

func EnterChan(chanID uint32, port int) {
	log.Printf("EnterChan chanID=%d, port=%d\n",chanID, port)
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: port}
	//dstAddr := &net.UDPAddr{IP: net.ParseIP("172.16.0.120"), Port: int(port)}
	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()

	/// 发送Subcribe
	sub  := &msg.Pack {}
	sub.Type = msg.Pack_SUBCRIBE
	sub.Subcribe = &msg.Pack_Subcribe {}
	sub.Subcribe.ChanID = chanID
	sub.Subcribe.ResouceID = "ABCD"
	sub.Subcribe.Start = 0
	sub.Subcribe.End = 128
	SendPack (sub, conn)

	time.Sleep (time.Second * 1)
	log.Println ("...")
	///等待SubscribeAck
	buffer := make([]byte, 1500)
	n, err := conn.Read (buffer)
	if nil != err {
		log.Panic(err)
	}
	subAck := &msg.Pack {}
	if err := proto.Unmarshal (buffer[:n], subAck); nil != err {
		log.Panic(err)
	}
	log.Println ("SubcribeAck", subAck)
}




func main() {
	chanID,port := SendReqChan ()
	time.Sleep (time.Second * 1)
	EnterChan (chanID, int(port))
}



