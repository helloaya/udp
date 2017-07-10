package main

import (
	"log"
	"net"
	"time"
	"udp/msg"
	"udp/file"
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
	file *file.File 
	sessinID uint32
}

func sendPack(p* msg.Pack, conn *net.UDPConn, remote *net.UDPAddr) {
	out,err := proto.Marshal (p)
	if nil != err {
		log.Panic(err)
	}
	conn.WriteTo (out, remote)
	log.Println ("Send", *p)
}


func handleSubscribe(c *Chan, p *msg.Pack, remote *net.UDPAddr, mgr *ChansManager) {
	if p.Subcribe.ChanID != c.ChanID {
		log.Println ("Invalid Subcribe", *p)
		return
	}

	if nil == c.file {
		newFile, err := mgr.reader.Read (p.Subcribe.ResouceID)
		if nil != err {
			/// 获取资源失败,忽略请求
			/// TODO 是否要拒绝
			log.Println (err)
			return
		}
		c.file = newFile
		c.sessinID = rand.Uint32()
	}

	if c.file.Id != p.Subcribe.ResouceID {
		///TODO 中止原来的session, 开启新Session
	}
	/// 响应SubcribeAck
	subAck := &msg.Pack {}
	subAck.Type = msg.Pack_SUBCRIBE_ACK
	subAck.SubcribeAck = &msg.Pack_SubcribeAck {}
	subAck.SubcribeAck.ResouceID = p.Subcribe.ResouceID
	subAck.SubcribeAck.SessionID = c.sessinID
	subAck.SubcribeAck.IsAccepted = true
	sendPack (subAck, c.Conn, remote)
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
			handleSubscribe (c, p, remote, mgr)
		case msg.Pack_REPORT:
			if STATUS_WAIT_REPORT == status {

			}
		default:
			log.Println ("Strange Pack", p.Type)
		}
	}

	///Chan完成
	mgr.clearChan (c.ClientID)
}
