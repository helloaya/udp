package core

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

const PACK_RECV_MAXTIMEOUT 		= time.Second * 10
const RATE_INIT					= 80 //初始80KB/s
const RATE_MIN					= 30 //最低30KB/s
const RATE_INC_STEP				= 20 //每次增长20KB
const RATE_DEC_STEP				= 10 //每次降低10KB
const RATE_THRESHOLD			= 70		

type Tunnel struct {
	TunnelID uint32
	ClientID uint32
	Conn *net.UDPConn
	SendIndex uint32
	Remote *net.UDPAddr
	SessionID uint32
	Bitmap *bitmap.Bitmap
	File *file.File
	SentPacks uint32
	Rate uint32 				//发送速率
	SentInterval time.Duration	//发送间隔
}


func handleSubscribe(tunnel *Tunnel, p *msg.Pack, remote *net.UDPAddr, mgr *TunnelManager) {
	if nil == p.Subcribe || p.Subcribe.TunnelID != tunnel.TunnelID {
		log.Println ("Invalid Subcribe", *p)
		return
	}

	if nil == tunnel.File {
		newFile, err := mgr.reader.Read (p.Subcribe.ResouceID)
		if nil != err {
			/// 获取资源失败,忽略请求
			/// TODO 要拒绝
			log.Println (err)
			return
		}
		if newFile.PiecesNum () <= p.Subcribe.End || p.Subcribe.Start > p.Subcribe.End {
			/// 无效访问,忽略请求
			/// TODO 要拒绝
			log.Printf ("Invalid Subscribe, invalid range[%d-%d]\n", p.Subcribe.Start,p.Subcribe.End)
			return
		}
		tunnel.File = newFile
		tunnel.SessionID = rand.Uint32()
		tunnel.Bitmap = bitmap.MakeBitmap(p.Subcribe.Start, p.Subcribe.End)
		tunnel.SendIndex  = p.Subcribe.Start
		tunnel.Remote = remote
		tunnel.SentPacks = 0
	}

	if tunnel.File.Id != p.Subcribe.ResouceID {
		///TODO 中止原来的session, 开启新Session
	}

	/// 响应SubcribeAck
	sendSubcribeAck(tunnel, p.Subcribe.ResouceID)
}


func handleReport(tunnel *Tunnel, p *msg.Pack, remote *net.UDPAddr, mgr *TunnelManager) bool {
	if nil == p.Report || nil == p.Report.Bitmap || tunnel.SessionID != p.Report.SessionID {
		log.Println ("Invalid Report", *p)
		return false
	}

	if nil == tunnel.Bitmap || nil == tunnel.Remote {
		log.Println ("Invalid Report, Tunnel Incomplete")
		return false
	}

	tunnel.Bitmap.Update (p.Report.Bitmap)


	/*
		1. 初始80KB
		2. 如果接收/发送速率比 < 0.7, 则降低10KB,最低30KB
		3. 如果接收/发送速率比 > 0.7, 则提高20KB
	*/
	
	/*
	ratio := p.Report.Rate * 100 / tunnel.Rate
	if RATE_THRESHOLD > ratio {
		tunnel.Rate -= RATE_DEC_STEP
		log.Println ("Decrease Rate = ", tunnel.Rate, " Ratio=", ratio)
	} else {
		tunnel.Rate += RATE_INC_STEP
		log.Println ("Increase Rate = ", tunnel.Rate, " Ratio=", ratio)
	}
	if RATE_MIN > tunnel.Rate {
		tunnel.Rate = RATE_MIN
	}
	*/
	if 0 < tunnel.SentPacks {
		ratio := p.Report.RecvedPacks * 100 / tunnel.SentPacks
		if RATE_THRESHOLD > ratio {
			//tunnel.Rate -= RATE_DEC_STEP
			tunnel.SentInterval += time.Millisecond
			log.Println ("SentInterval = ", tunnel.SentInterval, " Ratio=", ratio)
		} else {
			//tunnel.Rate += RATE_INC_STEP
			tunnel.SentInterval -= time.Millisecond * 2
			if time.Millisecond > tunnel.SentInterval {
				tunnel.SentInterval = time.Millisecond
			}
			log.Println ("SentInterval =", tunnel.SentInterval, " Ratio=", ratio)
		}
		if RATE_MIN > tunnel.Rate {
			tunnel.Rate = RATE_MIN
		}
	}
	return true
}

func handlePack(tunnel *Tunnel, p* msg.Pack, remote *net.UDPAddr, mgr *TunnelManager) bool {
	log.Println ("Recv",*p)
	/// 分发处理请求
	switch p.Type {
	case msg.Pack_SUBCRIBE:
		handleSubscribe (tunnel, p, remote, mgr)
	case msg.Pack_REPORT:
		handleReport (tunnel, p, remote, mgr)
	case msg.Pack_RELEASE:
		if p.Release.TunnelID == tunnel.TunnelID {
			///结束通道
			log.Println ("Release Channel=",tunnel.TunnelID)
			return false
		}
	default:
		log.Println ("Strange Pack", p.Type)
	}
	return true
}

func sendData(tunnel *Tunnel) {
	if nil != tunnel.Remote {
		sent := false
		for !sent {
			if !tunnel.Bitmap.Getbit(tunnel.SendIndex) {
				p := &msg.Pack {
					Type : msg.Pack_DATA,
					Data :  &msg.Pack_Data {
						SessionID : tunnel.SessionID,
						Index : tunnel.SendIndex,
						Payload : tunnel.File.GetPiece (tunnel.SendIndex),
					},
				}
				tunnel.Bitmap.Setbit(tunnel.SendIndex, true)
				tunnel.SentPacks += 1
				sendPack (p, tunnel.Conn, tunnel.Remote)
				sent = true
			}
			tunnel.SendIndex += 1
			if tunnel.SendIndex > tunnel.Bitmap.End {
				tunnel.SendIndex = tunnel.Bitmap.Start
				break
			}
		}
	}
}


func adjustRecvTimeout (tunnel *Tunnel) {
	//timeout := time.Second * time.Duration(file.SIZE_PIECE) / time.Duration(tunnel.Rate * 1024)
	expire := time.Now ().Add (tunnel.SentInterval)
	tunnel.Conn.SetReadDeadline (expire)
}

func (tunnel *Tunnel) Run(mgr *TunnelManager) {
	log.Printf ("Tunnel[%d] Running\n",tunnel.TunnelID)
	defer tunnel.Conn.Close ()

	timeout := time.Now()
	buffer := make([]byte, 1500)
	for {
		/// 计算读取超时,用于控制发送速率
		adjustRecvTimeout (tunnel)
		n, remote, err := tunnel.Conn.ReadFromUDP (buffer)
		if nil != err {
			nerr, ok := err.(net.Error)
			if !ok || (ok && !nerr.Timeout ()) {
				///其他错误,退出程序
				log.Panic (err)
			}
			///接收超时,释放Channel
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
