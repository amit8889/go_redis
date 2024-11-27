package main

import (
	"fmt"
	"log"
	"log/slog"
	"net"
	"runtime"
	"time"
)

const defaultListenAddr = ":5050"

type Config struct {
	ListenAdd string
}

type Server struct {
	Config    Config
	peers     map[*Peer]bool
	ln        net.Listener
	addPeerCh chan *Peer
	quitCh    chan struct{}
	msgCh     chan []byte
}

func NewServer(cfg Config) *Server {
	if len(cfg.ListenAdd) == 0 {
		cfg.ListenAdd = defaultListenAddr
	}
	return &Server{
		Config:    cfg,
		addPeerCh: make(chan *Peer),
		peers:     make(map[*Peer]bool),
		quitCh:    make(chan struct{}),
		msgCh:     make(chan []byte),
	}
}

func (s *Server) start() error {
	l, err := net.Listen("tcp", s.Config.ListenAdd)
	if err != nil {
		slog.Error("Error in tcp connection : ", "err", err)
		return err
	}
	defer l.Close()

	s.ln = l
	// redis pubsub channel
	go s.listenChannel()
	slog.Info("server is running on port:", "info", s.Config.ListenAdd)
	// connection accept
	return s.acceptConnection()
}
func (s *Server) listenChannel() {
	for {
		select {
		case rawMsg := <-s.msgCh:
			if err := s.handleRawMessage(rawMsg); err != nil {
				slog.Error("Error in handle raw message : ", "err", err)
			}
		case p := <-s.addPeerCh:
			s.peers[p] = true
		case <-s.quitCh:
			slog.Info("connect close")
			return
		default:
			// do nothing
			//slog.Info("No channne added")
			time.Sleep(500 * time.Millisecond)
		}
	}
}
func (s *Server) handleRawMessage(rawMsg []byte) error {
	// decode message
	fmt.Println(string(rawMsg))

	return nil

}
func (s *Server) acceptConnection() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("Error in accepting connection : ", "err", err)
			continue
		}
		// paraller processing
		go s.handleConnection(conn)

	}
}
func (s *Server) handleConnection(conn net.Conn) {
	// handle the connection
	peer := NewPeer(conn, s.msgCh)
	s.addPeerCh <- peer
	slog.Info("new peer connected ", "remoteAddress", conn.RemoteAddr())
	// reading loop
	if err := peer.readLoop(); err != nil {
		slog.Error("Error in reading loop : ", "err", err, "remoteAddress", conn.RemoteAddr())
	}
}

func main() {
	// start server
	runtime.GOMAXPROCS(runtime.NumCPU())
	server := NewServer(Config{ListenAdd: "localhost:8080"})
	log.Fatal(server.start())

}
