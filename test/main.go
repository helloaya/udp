package main

import (
	"log"
	"net"
	"time"
	"github.com/golang/protobuf/proto"
	"udp/msg"
	"udp/bitmap"
)

var (
	RESOURCE_ID = "1496830040.ts"
	RESOURCE_START uint32 = 0
	RESOURCE_END  uint32 = 960
	//SERVER_IP = "172.16.0.120"
	SERVER_IP = "127.0.0.1"
)

func SendReqChan() (uint32, uint32){
	conn, err := net.DialUDP("udp", 
						&net.UDPAddr{IP: net.IPv4zero, Port: 0}, 
						&net.UDPAddr{IP: net.ParseIP(SERVER_IP),
						Port: int(8888)})
	defer conn.Close()

	req := &msg.ReqChan {ClientID : 1024}
	out,err := proto.Marshal (req)
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

func SendReport (sessionID uint32, bits []byteconn *net.UDPConn) {

}

func RecvData (sessionID uint32, conn *net.UDPConn) {
	total := 0
	bits := bitmap.MakeBitmap (RESOURCE_START, RESOURCE_END)
	reportTick := time.Now ()
	for {
		expire := time.Now ().Add (time.Millisecond * 100)
		conn.SetReadDeadline (expire)

		buffer := make([]byte, 1500)
		n, _, err := conn.ReadFromUDP (buffer)
		if nil != err {
			nerr, ok := err.(net.Error)
			if !ok || (ok && !nerr.Timeout ()) {
				log.Panic (err)
			}
		} else {
			pack := &msg.Pack {}
			if err := proto.Unmarshal (buffer[:n], pack); nil != err {
				log.Panic(err)
			}
			total += 1
			bits.Setbit (pack.Data.Index, true)
			log.Println ("Recv", total, pack.Data.Index)
		}
		if time.Millisecond * 500 < time.Since (reportTick)  {
			reportTick = time.Now ()
			SendReport (conn);
		}
	}
}


func main() {
	chanID,port := SendReqChan ()
	conn, err := net.DialUDP("udp", 
								&net.UDPAddr{IP: net.IPv4zero, Port: 0}, 
								&net.UDPAddr{IP: net.ParseIP(SERVER_IP), 
								Port: int(port)})
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()
	sessionID := Subcribe (chanID, conn)
	RecvData (sessionID, conn)
}



