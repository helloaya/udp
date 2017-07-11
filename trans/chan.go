package main

import (
	"log"
	"net"
	"time"
	"udp/msg"
	"udp/file"
	"udp/bitmap"
	"math/rand"
	"github.com/golang/protobuf/proto"
)


const CHAN_RECV_TIMEOUT  = 10
const STATUS_WAIT_REPORT  = 1
const STATUS_WAIT_SUBSCRIBE  = 2


type Chan struct {
	ChanID uint32
	ClientID uint32
	Port uint16
	Conn *net.UDPConn
	
	SessionID uint32
	Bitmap *bitmap.Bitmap
	file *file.File 
	
}

func sendPack(p* msg.Pack, conn *net.UDPConn, remote *net.UDPAddr) {
	out,err := proto.Marshal (p)
	if nil != err {
		log.Panic(err)
	}
	conn.WriteTo (out, remote)
	//log.Println ("Send", *p)
}


func handleSubscribe(c *Chan, p *msg.Pack, remote *net.UDPAddr, mgr *ChansManager) bool {
	if p.Subcribe.ChanID != c.ChanID {
		log.Println ("Invalid Subcribe", *p)
		return false
	}

	if nil == c.file {
		newFile, err := mgr.reader.Read (p.Subcribe.ResouceID)
		if nil != err {
			/// 获取资源失败,忽略请求
			/// TODO 是否要拒绝
			log.Println (err)
			return false
		}
		c.file = newFile
		c.SessionID = rand.Uint32()
		c.Bitmap = bitmap.MakeBitmap(p.Subcribe.Start, p.Subcribe.End)
	}

	if c.file.Id != p.Subcribe.ResouceID {
		///TODO 中止原来的session, 开启新Session
	}

	/// 响应SubcribeAck
	subAck := &msg.Pack {}
	subAck.Type = msg.Pack_SUBCRIBE_ACK
	subAck.SubcribeAck = &msg.Pack_SubcribeAck {}
	subAck.SubcribeAck.ResouceID = p.Subcribe.ResouceID
	subAck.SubcribeAck.SessionID = c.SessionID
	subAck.SubcribeAck.IsAccepted = true
	sendPack (subAck, c.Conn, remote)
	return false
}


func handleReport(c *Chan, p *msg.Pack, remote *net.UDPAddr, mgr *ChansManager) bool {
	if c.SessionID != p.Report.SessionID {
		log.Println ("Invalid Report", *p)
		return false
	}
	c.Bitmap.Update (p.Report.Bitmap)
	return true
}

func sendData(c *Chan,  remote *net.UDPAddr) {
	///TODO 优化算法,灵活调节下面各个参数
	rate := uint32(30)   //30KB/s
	duration := uint32(1000) //发送1000毫秒
	packs := (rate * 1024 / file.SIZE_PIECE) * (duration / 1000)  //估算要发送多少包
	interval := time.Second / time.Duration(rate * 1024 / file.SIZE_PIECE) //估算发送包的间隔
	index := c.Bitmap.Start

	log.Printf ("Send Data: Rate=%d,Packs=%d, Interval=%d\n",
			rate,packs,interval)
	for 0 < packs {
		p := &msg.Pack {}
		p.Type = msg.Pack_DATA
		p.Data = &msg.Pack_Data {}
		p.Data.SessionID = c.SessionID
		p.Data.Index = index
		p.Data.Payload = c.file.GetPiece (index)
		sendPack (p, c.Conn, remote)
		index += 1
		if index > c.Bitmap.End {
			index = c.Bitmap.Start
		}
		packs -= 1
		time.Sleep (interval)
	}
}


func (c *Chan) Run(mgr *ChansManager) {
	defer c.Conn.Close ()
	log.Printf ("Chan[%d] Running\n",c.ChanID)
	status := STATUS_WAIT_SUBSCRIBE
	buffer := make([]byte, 1500)
	for {
		expire := time.Now ().Add (time.Second * CHAN_RECV_TIMEOUT)
		c.Conn.SetReadDeadline (expire)
		n, remote, err := c.Conn.ReadFromUDP (buffer)
		if nil != err {
			nerr, ok := err.(net.Error)
			if ok && nerr.Timeout () {
				///超时
				log.Printf ("Chanl[%d] Timeout, Exit\n", c.ChanID)
				break
			} else {
				///其他错误,退出程序
				log.Panic (err)
			}
		}
		p := &msg.Pack {}
		if err := proto.Unmarshal (buffer[:n], p); nil != err {
			//忽略这个错误
			log.Println(err)
			continue
		}

		/// 分发处理请求
		switch p.Type {
		case msg.Pack_SUBCRIBE:
			if ok := handleSubscribe (c, p, remote, mgr); ok {
				status = STATUS_WAIT_REPORT
			}
		case msg.Pack_REPORT:
			if STATUS_WAIT_REPORT == status {
				handleReport (c, p, remote, mgr)
			}
		case msg.Pack_RELEASE:
			if p.Release.ChanID == c.ChanID {
				///结束通道
				log.Println ("End of Chan",c.ChanID)
				break
			}
		default:
			log.Println ("Strange Pack", p.Type)
		}

		/// 发送数据
		sendData (c, remote)
	}

	///Chan完成
	mgr.clearChan (c.ClientID)
}

