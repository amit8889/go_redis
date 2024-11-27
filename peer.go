package main

import (
	"fmt"
	"net"
)

type Peer struct {
	conn net.Conn
}

func NewPeer(conn net.Conn) *Peer {
	return &Peer{conn: conn}
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

	}

}
