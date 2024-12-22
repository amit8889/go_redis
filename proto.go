package main

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"

	"github.com/tidwall/resp"
)

const (
	CommandSET = "SET"
	CommandGET = "GET"
)

type Commond interface {
}

type SetCommand struct {
	key, val string
}
type GetCommand struct {
	key string
}

func parseCommond(raw string) (Commond, error) {
	// parse message
	//fmt.Println("OK", raw)
	rd := resp.NewReader(bytes.NewBufferString(raw))
	for {
		v, _, err := rd.ReadValue()
		//fmt.Println(v)
		if err == io.EOF {
			break
		}
		if err != nil {
			//fmt.Println(err)
			return nil, fmt.Errorf("invalid syntax")
		}
		if v.Type() == resp.Array && len(v.Array()) > 0 {
			//	fmt.Printf("  #%d, value: '%s'\n", v.Type(), v)
			switch v.Array()[0].String() {
			case CommandSET:
				if len(v.Array()) != 3 {
					return nil, fmt.Errorf("invalid set commend")
				}
				set := SetCommand{}
				set.key = v.Array()[1].String()
				set.val = v.Array()[2].String()
				return set, nil
			case CommandGET:
				if len(v.Array()) != 2 {
					return nil, fmt.Errorf("invalid get commend")
				}
				get := GetCommand{}
				get.key = v.Array()[1].String()
				return get, nil
			default:
				{
					slog.Info("Invalid command send", "INFO", v.Array()[0].String())
					return nil, fmt.Errorf("unknown commend")
				}
			}
		}
	}
	return fmt.Errorf("EOF"), nil
}
