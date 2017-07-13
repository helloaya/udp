package core

import (
	"log"
	"net"
	"time"
	"udp/msg"
	"github.com/golang/protobuf/proto"
)

const PACK_RECV_MAXTIMEOUT 	= time.Second * 10





type Tunnel struct {
	TunnelID uint32
	ClientID uint32
	Conn *net.UDPConn
	Session *TunnelSession
}


func handleSubscribe(tunnel *Tunnel, p *msg.Pack, clientAddr *net.UDPAddr, mgr *TunnelManager) {
	sub := p.Subcribe
	if nil == sub || sub.TunnelID != tunnel.TunnelID {
		log.Println ("Invalid Subcribe", *p)
		return
	}
	/// 创建新Session
	if nil == tunnel.Session || tunnel.Session.getResourceID() != sub.ResouceID {
		tunnel.Session = makeTunnelSession (sub,  mgr.reader, clientAddr)
		if nil == tunnel.Session {
			///TO DO创建Session失败,拒绝
			return
		}
	}
	/// 响应SubcribeAck
	sendSubcribeAck(tunnel)
}


func handleReport(tunnel *Tunnel, p *msg.Pack, mgr *TunnelManager) {
	if nil == tunnel.Session {
		log.Println ("Invalid report, Session not set up")
		return
	}

	report := p.Report
	if nil == report || nil == report.Bitmap || tunnel.Session.ID != report.SessionID {
		log.Println ("Invalid Report", *p)
		return
	}
	// 更新客户端接收记录,调整发送间隔
	tunnel.Session.update (report)
}

func handlePack(tunnel *Tunnel, pack* msg.Pack, clientAddr *net.UDPAddr, mgr *TunnelManager) bool {
	log.Println ("Recv",*pack)
	/// 分发处理请求
	switch pack.Type {
	case msg.Pack_SUBCRIBE:
		handleSubscribe (tunnel, pack, clientAddr, mgr)
	case msg.Pack_REPORT:
		handleReport (tunnel, pack, mgr)
	case msg.Pack_RELEASE:
		if nil != pack.Release && pack.Release.TunnelID == tunnel.TunnelID {
			///结束通道
			log.Println ("Release Tunnel =",tunnel.TunnelID)
			return false
		}
	default:
		log.Println ("Strange Pack", *pack)
	}
	return true
}


func setReadTimeout(tunnel *Tunnel) {
	var expire time.Time
	if nil != tunnel.Session {
		expire = time.Now ().Add(tunnel.Session.SendInterval)
	} else {
		expire = time.Now ().Add(time.Millisecond * 10)
	}
	tunnel.Conn.SetReadDeadline (expire)
}


/// 收收收,发发发
func (tunnel *Tunnel) Run(mgr *TunnelManager) {
	log.Printf ("Tunnel[%d] Running\n",tunnel.TunnelID)
	defer tunnel.Conn.Close ()

	timeout := time.Now()
	buffer := make([]byte, 1500) //TODO,固定1500字节接收缓冲是否妥?
	for {
		/// 调整读取超时,用于控制发送速率
		setReadTimeout (tunnel)
		n, remote, err := tunnel.Conn.ReadFromUDP (buffer)
		if nil != err {
			nerr, ok := err.(net.Error)
			if !ok || (ok && !nerr.Timeout ()) {
				///其他错误,退出Tunnel
				log.Println (err)
				break
			}
			///接收超时,释放Tunnel
			if PACK_RECV_MAXTIMEOUT < time.Since(timeout) {
				log.Printf ("Tunnel[%d] Timeout, Exit\n", tunnel.TunnelID)
				break
			}
		} else {
			p := &msg.Pack {}
			if err := proto.Unmarshal (buffer[:n], p); nil != err {
				//忽略解析错误
				log.Println(err)
				continue
			}
			if !handlePack (tunnel, p, remote, mgr) {
				break
			}
			timeout = time.Now ()
		}
		/// 发送数据
		sendData (tunnel)
	}
	///Tunnel完成
	mgr.delete (tunnel)
}
