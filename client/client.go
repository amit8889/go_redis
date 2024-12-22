package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/tidwall/resp"
)

type Client struct {
	addr string
}

type RedisClient struct {
	conn net.Conn
}

func NewClient() *Client {
	return &Client{
		addr: "/tmp/myredis.sock",
	}
}

func NewRedisClient(connection net.Conn) *RedisClient {
	return &RedisClient{conn: connection}
}

func (c *Client) Connect() {
	conn, err := net.Dial("unix", c.addr)
	if err != nil {
		slog.Error("Connection error", "ERROR", err.Error())
		return
	}
	redisConn := NewRedisClient(conn)
	defer conn.Close()
	slog.Info("Connected successfully !!!!")

	go redisConn.HandleCommand()
	redisConn.ReadMessage()
}

func (r *RedisClient) HandleCommand() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("redis-cli> ")
		if !scanner.Scan() {
			break
		}
		cmd := strings.Split(scanner.Text(), " ")

		if strings.ToLower(cmd[0]) == "exit" {
			r.conn.Close()
			return
		}
		var buf bytes.Buffer
		wr := resp.NewWriter(&buf)
		switch strings.ToUpper(cmd[0]) {
		case "SET":
			if len(cmd) != 3 {
				fmt.Println("Invalid command")
				continue
			}
			err := wr.WriteArray([]resp.Value{
				resp.StringValue("SET"),
				resp.StringValue(cmd[1]),
				resp.StringValue(cmd[2]),
			})
			if err != nil {
				fmt.Printf("Error sending SET command: %v\n", err)
				continue
			}
		case "GET":
			if len(cmd) != 2 {
				fmt.Println("Invalid command")
				continue
			}
			err := wr.WriteArray([]resp.Value{
				resp.StringValue("GET"),
				resp.StringValue(cmd[1]),
			})
			if err != nil {
				fmt.Printf("Error sending GET command: %v\n", err)
				continue
			}
		default:
			fmt.Println("Unknown command")
			continue
		}
		_, err := r.ExecuteCommand(buf)
		if err != nil {
			fmt.Printf("Error executing command: %v\n", err)
			continue
		}

	}
}

func (r *RedisClient) ExecuteCommand(buf bytes.Buffer) (string, error) {
	//make it buffer
	_, err := r.conn.Write(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}
	return "", nil
}

func (r *RedisClient) ReadMessage() {
	reader := bufio.NewReader(r.conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			reConnClosed := regexp.MustCompile(`use of closed network connection`)
			if reConnClosed.MatchString(err.Error()) {
				fmt.Println("Connection Exit")
				return
			}
			fmt.Printf("Error reading message: %v\n", err)
			return
		}
		// Process the received message
		fmt.Print(msg)
		fmt.Print("redis-cli> ")
	}
}

func main() {
	client := NewClient()
	client.Connect()
}
