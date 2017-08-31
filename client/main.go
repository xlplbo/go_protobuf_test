package main

import (
	"fmt"
	"log"
	"net"

	"github.com/golang/protobuf/proto"
	"github.com/xlplbo/go_protobuf_test/protocol"
)

func senddata(c net.Conn, serial protocol.C2SCmd, msg proto.Message) error {
	buff, err := protocol.Pack(int32(serial), msg)
	if err != nil {
		return err
	}
	if _, err := c.Write(buff); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func main() {
	l, err := net.Dial("tcp", "0.0.0.0:7788")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for i := 0; ; i++ {
		if err := senddata(l, protocol.C2SCmd_Login, &protocol.C2SLogin{
			User:   "zhangsan",
			Passwd: "zhangsan",
		}); err != nil {
			break
		}

		if err := senddata(l, protocol.C2SCmd_Chat, &protocol.C2SChat{
			Count:   uint32(i),
			Context: "hello",
		}); err != nil {
			break
		}
		fmt.Println(i)
	}
}
