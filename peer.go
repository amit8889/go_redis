package main

import (
	"fmt"
	"net"
)

type Peer struct {
	conn  net.Conn
	msgCh chan []byte
}

func NewPeer(conn net.Conn, msgCh chan []byte) *Peer {
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
		fmt.Println(string(buf[:n]))
		msgBuf := make([]byte, n)
		copy(msgBuf, buf[:n])
		p.msgCh <- msgBuf

	}

}
