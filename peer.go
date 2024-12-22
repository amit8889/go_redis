package main

import (
	"log/slog"
	"net"
)

type Peer struct {
	conn  net.Conn
	msgCh chan Message
}

func NewPeer(conn net.Conn, msgCh chan Message) *Peer {
	return &Peer{conn: conn, msgCh: msgCh}
}
func (p *Peer) readLoop() error {
	for {
		buf := make([]byte, 1024)
		n, err := p.conn.Read(buf)
		if err != nil {
			return err
		}
		// do something with the data
		//fmt.Println(string(buf[:n]))
		msgBuf := make([]byte, n)
		copy(msgBuf, buf[:n])
		msg := Message{
			peer: p,
			data: buf[:n],
		}
		p.msgCh <- msg

	}

}

func (p *Peer) SendMessage(msg string) {
	_, err := p.conn.Write([]byte(msg + "\n"))
	if err != nil {
		slog.Info("Error in writeing peer message", "ERROR", err.Error())
	}
}
