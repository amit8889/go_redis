package main

import (
	"log"
	"log/slog"
	"net"
	"os"
	"regexp"
	"runtime"
	"time"
)

const defaultListenAddr = ":5050"

type Config struct {
	ListenAdd string
}
type Message struct {
	peer *Peer
	data []byte
}

type Server struct {
	Config    Config
	peers     map[*Peer]bool
	ln        net.Listener
	addPeerCh chan *Peer
	quitCh    chan struct{}
	msgCh     chan Message
	kv        *KV
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
		msgCh:     make(chan Message),
		kv:        NewKV(),
	}
}

func (s *Server) start() error {
	socketPath := s.Config.ListenAdd

	// Remove existing socket file if it exists
	if _, err := os.Stat(socketPath); err == nil {
		os.Remove(socketPath)
	}
	// Create a Unix domain socket listener
	l, err := net.ListenUnix("unix", &net.UnixAddr{Name: socketPath, Net: "unix"})
	if err != nil {
		slog.Error("Error creating Unix socket:", "err", err)
		return nil
	}
	defer os.Remove(socketPath)
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
		case msg := <-s.msgCh:
			if err := s.handleMessage(msg); err != nil {
				//	slog.Error("Error in handle raw message : ", "err", err)
			}
		case p := <-s.addPeerCh:
			s.peers[p] = true
		case <-s.quitCh:
			slog.Info("connect close")
			return
		default:
			// do nothing
			time.Sleep(500 * time.Millisecond)
		}
	}
}
func (s *Server) handleMessage(msg Message) error {
	cmd, err := parseCommond(string(msg.data))
	if err != nil {
		msg.peer.SendMessage("Invalid command")
		return err
	}
	switch v := cmd.(type) {
	case SetCommand:
		slog.Info("SET command ", "INFO", "KEY : "+v.key, "val:", v.val)
		s.kv.Set(v.key, v.val)
		msg.peer.SendMessage("OK")
		break
	case GetCommand:
		slog.Info("GET command ", "INFO", "KEY : "+v.key)
		val := s.kv.Get(v.key)
		if len(val) == 0 {
			msg.peer.SendMessage("nil")
		} else {
			msg.peer.SendMessage(string(val))
		}

		break
	default:
		slog.Info("unknown command")
		msg.peer.SendMessage("Invalid command")
	}

	return nil
}

func (s *Server) acceptConnection() error {
	for {
		conn, err := s.ln.Accept()
		//conn.SetDeadline(time.Now().Add(2 * time.Second))
		if err != nil {
			slog.Error("Error in accepting connection : ", "err", err)
			continue
		}
		if err = conn.SetDeadline(time.Now().Add(60 * 60 * time.Second)); err != nil {
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
		reConnClosed := regexp.MustCompile(`i/o timeout`)
		if reConnClosed.MatchString(err.Error()) {
			//slog.Info("connection closed due to timeout")
			return
		}
		slog.Error("Error in reading loop : ", "err", err, "remoteAddress", conn.RemoteAddr())
	}

}

func main() {
	// start server
	runtime.GOMAXPROCS(runtime.NumCPU())
	server := NewServer(Config{ListenAdd: "/tmp/myredis.sock"})
	//server := NewServer(Config{})
	log.Fatal(server.start())

}
