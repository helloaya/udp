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

const PACK_RECV_MAXTIMEOUT = time.Second * 2
const PACK_RECV_TIMEOUT  = time.Millisecond * 100 
const STATUS_WAIT_REPORT  = 1
const STATUS_WAIT_SUBSCRIBE  = 2


type Chan struct {
	ChanID uint32
	ClientID uint32
	Port uint16
	Conn *net.UDPConn
	Status int
	SendIndex uint32
	Remote *net.UDPAddr

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
	if p.Type != msg.Pack_DATA {
		log.Println ("Send", *p)
	} else {
		log.Println ("Send Data, index=", p.Data.Index, *remote)
	}
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
			/// TODO 要拒绝
			log.Println (err)
			return false
		}
		if newFile.PiecesNum () <= p.Subcribe.End || p.Subcribe.Start > p.Subcribe.End {
			/// 无效访问,忽略请求
			/// TODO 要拒绝
			log.Printf ("Invalid Subscribe, invalid range[%d-%d]\n", p.Subcribe.Start,p.Subcribe.End)
			return false
		}

		c.file = newFile
		c.SessionID = rand.Uint32()
		c.Bitmap = bitmap.MakeBitmap(p.Subcribe.Start, p.Subcribe.End)
		c.SendIndex  = p.Subcribe.Start
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
	return true
}


func handleReport(c *Chan, p *msg.Pack, remote *net.UDPAddr, mgr *ChansManager) bool {
	if c.SessionID != p.Report.SessionID {
		log.Println ("Invalid Report", *p)
		return false
	}
	c.Bitmap.Update (p.Report.Bitmap)
	return true
}




func displayPack (p *msg.Pack) {
	if p.Type == msg.Pack_REPORT {
		log.Println ("Report, bits=", p.Report.Bitmap)
	} else {
		log.Println ("Recv", *p)
	}
}

func handlePack(c *Chan, p* msg.Pack, remote *net.UDPAddr, mgr *ChansManager) bool {
	displayPack (p)

	/// 分发处理请求
	switch p.Type {
	case msg.Pack_SUBCRIBE:
		if ok := handleSubscribe (c, p, remote, mgr); ok {
			c.Status = STATUS_WAIT_REPORT
			c.Remote = remote
		}
	case msg.Pack_REPORT:
		if STATUS_WAIT_REPORT == c.Status {
			handleReport (c, p, remote, mgr)
		}
	case msg.Pack_RELEASE:
		if p.Release.ChanID == c.ChanID {
			///结束通道
			log.Println ("Release Channel=",c.ChanID)
			return false
		}
	default:
		log.Println ("Strange Pack", p.Type)
	}
	return true
}

func sendData(c *Chan) {
	if nil != c.Remote && !c.Bitmap.IsComplete(){
		for {
			if !c.Bitmap.Getbit(c.SendIndex) {
				p := &msg.Pack {}
				p.Type = msg.Pack_DATA
				p.Data = &msg.Pack_Data {}
				p.Data.SessionID = c.SessionID
				p.Data.Index = c.SendIndex
				p.Data.Payload = c.file.GetPiece (c.SendIndex)
				sendPack (p, c.Conn, c.Remote)
				c.SendIndex += 1
				break
			}
			c.SendIndex += 1
			if c.SendIndex > c.Bitmap.End {
				c.SendIndex = c.Bitmap.Start
			}
		}
	}
}

func (c *Chan) Run(mgr *ChansManager) {
	log.Printf ("Chan[%d] Running\n",c.ChanID)
	defer c.Conn.Close ()

	timeout := time.Now()
	buffer := make([]byte, 1500)
	for {
		expire := time.Now ().Add (PACK_RECV_TIMEOUT)
		c.Conn.SetReadDeadline (expire)
		n, remote, err := c.Conn.ReadFromUDP (buffer)
		if nil != err {
			nerr, ok := err.(net.Error)
			if !ok || (ok && !nerr.Timeout ()) {
				///其他错误,退出程序
				log.Panic (err)
			}

			///接收超时,释放Channel
			if PACK_RECV_MAXTIMEOUT < time.Since(timeout) {
				log.Printf ("Chan[%d] Timeout, Exit\n", c.ChanID)
				break
			}
		} else {
			p := &msg.Pack {}
			if err := proto.Unmarshal (buffer[:n], p); nil != err {
				//忽略解析错误
				log.Println(err)
				continue
			}
			
			if !handlePack (c, p, remote, mgr) {
				break
			}
			timeout = time.Now ()
		}

		/// 发送数据
		sendData (c)
	}

	///Chan完成
	mgr.clearChan (c.ClientID)
}

