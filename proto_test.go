package main

import (
	"fmt"
	"log"
	"testing"
)

func TestProtocol(t *testing.T) {
	raw := "*3\r\n$3\r\nSET\r\n$5\r\nmykey\r\n$7\r\nmyvalue\r\n"
	cmd, err := parseCommond(raw)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cmd)

}
