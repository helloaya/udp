package main

import (
	"log"
	"net"
	"time"
	"github.com/golang/protobuf/proto"
	"crypto/md5"
	"math/rand"
	"encoding/hex"
	//"io/ioutil"
	"udp/msg"
	"udp/bitmap"
	"udp/file"
)

var (
	RESOURCE_ID = "1496830040.ts"
	RESOURCE_START uint32 = 0
	RESOURCE_END  uint32 = 959
	RESOURCE_LENGTH uint32 = 1342696
	//SERVER_IP = "172.16.0.120"
	SERVER_IP = "127.0.0.1"
	//SERVER_IP = "142.234.27.42"
	//SERVER_IP = "147.255.237.34"
	SEND_REPORT_INTERVAL = time.Millisecond * 1000
)

func SendReqChan() (uint32, uint32){
	conn, err := net.DialUDP("udp", 
						&net.UDPAddr{IP: net.IPv4zero, Port: 0}, 
						&net.UDPAddr{IP: net.ParseIP(SERVER_IP),
						Port: int(8888)})
	defer conn.Close()

	req := &msg.ReqTunnel {ClientID : rand.Uint32()}
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

	resp := &msg.ReqTunnelAck {}
	if err := proto.Unmarshal (buffer[:n], resp); nil != err {
		log.Panic(err)
	}
	return resp.TunnelID, resp.TunnelPort
}

func displayPack (p *msg.Pack) {
	log.Println ("Send  ", *p)
}

func SendPack(p* msg.Pack, conn *net.UDPConn) {
	out,err := proto.Marshal (p)
	if nil != err {
		log.Panic(err)
	}

	conn.Write (out)
	displayPack(p)
}

func Subcribe(chanID uint32, conn *net.UDPConn) uint32{
	log.Printf("Subcribe chanID=%dn",chanID)

	/// 发送Subcribe
	sub  := &msg.Pack {}
	sub.Type = msg.Pack_SUBCRIBE
	sub.Subcribe = &msg.Pack_Subcribe {}
	sub.Subcribe.TunnelID = chanID
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

func SendReport(sessionID uint32, bits []byte, rate uint32, totalPacks uint32, conn *net.UDPConn) {
	report := &msg.Pack {
		Type : msg.Pack_REPORT,
		Report : &msg.Pack_Report {
				SessionID : sessionID,
				Rate : rate,
				RecvedPacks : totalPacks,
				Bitmap : bits,
			},
	}

	SendPack (report, conn)
}

func SendRelease(chanID uint32, conn *net.UDPConn) {
	report := &msg.Pack {}
	report.Type = msg.Pack_RELEASE
	report.Release = &msg.Pack_Release {}
	report.Release.TunnelID = chanID
	SendPack (report, conn)
	time.Sleep (time.Second)
}

func RecvData(sessionID uint32, conn *net.UDPConn) {
	data := make([]byte, RESOURCE_LENGTH)
	bits := bitmap.MakeBitmap (RESOURCE_START, RESOURCE_END)
	reportTick := time.Now ()
	totalRecv := 0
	totalPacks := 0
	startTick := time.Now ()
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
			totalPacks += 1
			if bits.Getbit (pack.Data.Index) {
				log.Println ("Recv repeat pack=", pack.Data.Index)
			} else {
				bits.Setbit (pack.Data.Index, true)
				copy (data[(pack.Data.Index - RESOURCE_START) * file.SIZE_PIECE:], pack.Data.Payload)
				totalRecv += len (pack.Data.Payload)
			}
		}
		if SEND_REPORT_INTERVAL < time.Since (reportTick)  {
			reportTick = time.Now ()
			rate := 0
			d := int(time.Since (startTick) / time.Millisecond)
			if 0 < d {
				rate = totalRecv / d
				log.Println ("Rate=", rate, "KB/s")
			}
			SendReport (sessionID, bits.Get(), uint32(rate),  uint32(totalPacks), conn);
		}

		if bits.IsComplete () {
			log.Println ("Recv data complete")

		 	md5Ctx := md5.New()
		    md5Ctx.Write(data)
		    cipherStr := md5Ctx.Sum(nil)
		    log.Println (hex.EncodeToString(cipherStr))

		    ///ioutil.WriteFile ("recv.ts", data,0777)
			break
		}
	}
}


func main() {
	rand.Seed (int64(time.Now ().Nanosecond()))
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
	RecvData(sessionID, conn)
	SendRelease(chanID, conn)
}



