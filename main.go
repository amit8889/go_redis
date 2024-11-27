package main

import (
	"log"
	"log/slog"
	"net"
	"runtime"
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
	}
}

func (s *Server) start() error {
	l, err := net.Listen("tcp", s.Config.ListenAdd)
	if err != nil {
		slog.Error("Error in tcp connection : ", "err", err)
		return err
	}
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
		case p := <-s.addPeerCh:
			s.peers[p] = true
		case <-s.quitCh:
			slog.Info("connect close")
			return
		default:
			// do nothing
			//slog.Info("No channne added")
		}
	}
}
func (s *Server) acceptConnection() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("Error in accepting connection : ", "err", err)
			continue
		}
		go s.handleConnection(conn)

	}
}
func (s *Server) handleConnection(conn net.Conn) {
	// handle the connection
	peer := NewPeer(conn)
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
