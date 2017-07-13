package core


import (
	"udp/file"
	"udp/msg"
	"udp/bitmap"
	"net"
	"time"
	"log"
	"math/rand"
)

const DEFAULT_SEND_INTERVAL  = time.Millisecond * 10
const SEND_INTERVAL_INC_STEP = time.Millisecond
const SEND_INTERVAL_DEC_STEP = time.Millisecond * 2
const MIN_SEND_INTERVAL		 = time.Millisecond
const MAX_SEND_INTERVAL		 = time.Millisecond * 20
const MAX_LOSS_RATIO		 = 30
const INIT_LAST_SENT		 = ^uint32(0)

type TunnelSession struct {
	ID 				uint32 //Session ID
	LastSent		uint32 //记录上一个发送的包
	TotalSentPacks 	uint32 //总发送包数
	TotalSentBytes  uint32 //总发送字节数
	Rate 			uint32 //发送速率. KB/s
	SendInterval 	time.Duration
	RecvRecord 		*bitmap.Bitmap  //接收记录
	File 			*file.File
	ClientAddr 		*net.UDPAddr
	StartTime		time.Time
}

func (session *TunnelSession) getResourceID() string {
	return session.File.ID
}

func makeTunnelSession (subscibe *msg.Pack_Subcribe, reader file.Reader, clientAddr *net.UDPAddr) *TunnelSession {
	newFile, err := reader.Read (subscibe.ResouceID)
	if nil != err {
		log.Println (err)
		return nil
	}
	if newFile.PiecesNum () <= subscibe.End || subscibe.Start > subscibe.End {
		log.Printf ("Invalid Subscribe, range[%d-%d]\n", subscibe.Start, subscibe.End)
		return nil
	}

	record := bitmap.MakeBitmap(subscibe.Start, subscibe.End)
	session := &TunnelSession {
		ID : rand.Uint32(),
		LastSent : INIT_LAST_SENT,
		TotalSentPacks : 0,
		TotalSentBytes : 0,
		Rate : 0,
		SendInterval : DEFAULT_SEND_INTERVAL,
		RecvRecord : record,
		File : newFile,
		ClientAddr : clientAddr,
		StartTime : time.Now(),
	}
	return session
}

func (session *TunnelSession) update(report *msg.Pack_Report) {
	session.RecvRecord.Update(report.Bitmap)
	/*
		发送间隔调整策略
		1.默认DEFAULT_SEND_INTERVAL (10ms)
		2.如果总丢包率 > MAX_LOSS_RATIO (30%), 则发送间隔增加SEND_INTERVAL_INC_STEP(1ms)
		3.如果总丢包率 < MAX_LOSS_RATIO (30%), 则发送间隔减少SEND_INTERVAL_DEC_STEP(2ms), 最少MIN_SEND_INTERVAL
	 */
	if 0 < session.TotalSentPacks {
		ratio := report.RecvedPacks * 100 / session.TotalSentPacks
		if ratio > 100 {
			ratio = 100
		}
		if (100 - MAX_LOSS_RATIO) > ratio {
			session.SendInterval += SEND_INTERVAL_INC_STEP
			//log.Println ("SendInterval = ", session.SendInterval, " Ratio=", ratio)
		} else {
			session.SendInterval -= SEND_INTERVAL_DEC_STEP
			if 0 >= session.SendInterval {
				session.SendInterval = MIN_SEND_INTERVAL
			}

			if MAX_SEND_INTERVAL < session.SendInterval {
				session.SendInterval = MAX_SEND_INTERVAL
			}
			//log.Println ("SendInterval =", session.SendInterval, " Ratio=", ratio)
		}
		log.Printf("Session[%d] Packs[%d] Bytes[%d] Rate[%d KB/s] Interval[%s] Loss[%d]\n",
			session.ID, session.TotalSentPacks, session.TotalSentBytes, session.Rate, session.SendInterval, 100 - ratio)
	}
}

func (session *TunnelSession) getNextPack() (uint32,[]byte, bool) {
	var payload []byte
	var index uint32 = INIT_LAST_SENT
	var got bool = false
	for !got {
		if INIT_LAST_SENT == session.LastSent {
			session.LastSent = session.RecvRecord.Start
		} else {
			session.LastSent += 1
			if session.LastSent > session.RecvRecord.End {
				session.LastSent = INIT_LAST_SENT
				break
			}
		}
		if !session.RecvRecord.Getbit(session.LastSent) {
			index = session.LastSent
			payload = session.File.GetPiece (index)
			got = true
			session.RecvRecord.Setbit(index, true)
			session.TotalSentPacks += 1
			session.TotalSentBytes += uint32(len(payload))

			///更新发送速率
			duration := uint32(time.Since(session.StartTime)/time.Millisecond)
			if 0 < duration {
				session.Rate = session.TotalSentBytes /duration
			}
		}
	}
	return index, payload, got
}
