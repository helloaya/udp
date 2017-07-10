package main

import (
	"log"
	"net"
	"time"
	//"udp/resource"
	"udp/msg"
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
}



func doSubscribe(conn *net.UDPConn,p *msg.Pack) {
	log.Printf ("Subcribe:[%s][%d - %d]\n", p.Subcribe.ResouceID, p.Subcribe.Start, p.Subcribe.End)

}


func (c *Chan) Run(mgr *ChansManager) {
	log.Printf ("Chan[%d] Running\n",c.ChanID)


	status := STATUS_WAIT_SUBSCRIBE
	buffer := make([]byte, 1500)
	for {
		expire := time.Now ().Add (time.Second * CHAN_RECV_TIMEOUT)
		c.Conn.SetReadDeadline (expire)
		n, _, err := c.Conn.ReadFromUDP (buffer)
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

		switch p.Type {
		case msg.Pack_SUBCRIBE:
			if p.Subcribe.ChanID == c.ChanID {
				doSubscribe(c.Conn, p)
			}
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
