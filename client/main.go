package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/xlplbo/go_protobuf_test/protocol"
)

var handles map[int32]func([]byte)
var chStop chan error

func init() {
	chStop = make(chan error, 1)
	handles = make(map[int32]func([]byte))
	registerHandle(protocol.S2CCmd_Invalid, stopClient)
	registerHandle(protocol.S2CCmd_Result, showMsg)
}

func registerHandle(id protocol.S2CCmd, f func([]byte)) {
	nID := int32(id)
	if _, ok := handles[nID]; ok {
		log.Printf("protocol(%d) handle repeat\n", nID)
		return
	}
	handles[nID] = f
}

func stopClient(msg []byte) {
	chStop <- fmt.Errorf("data invalid")
}

func showMsg(msg []byte) {
	var result protocol.S2CResult
	if err := proto.Unmarshal(msg, &result); err != nil {
		log.Println(err)
		return
	}
	log.Println(result.Context)
}

func main() {
	chSig := make(chan os.Signal, 1)
	signal.Notify(chSig, os.Interrupt)
	go func(sig <-chan os.Signal, stop chan<- error) {
		for {
			select {
			case s := <-sig:
				stop <- fmt.Errorf("stop clinet: %s", s.String())
				return
			}
		}
	}(chSig, chStop)

	chConn := make(chan net.Conn, 1)
	go func(ch chan<- net.Conn) {
		for {
			select {
			case <-time.Tick(time.Second):
				log.Println("connect server...")
				conn, err := net.Dial("tcp", "127.0.0.1:7788")
				if err != nil {
					log.Println(err)
					continue
				}
				ch <- conn
				return
			}
		}
	}(chConn)

	conn := <-chConn
	defer conn.Close()
	log.Printf("%s established", conn.RemoteAddr().String())

	go func(conn net.Conn, stop chan<- error) {
		var data []byte
		buff := make([]byte, protocol.MaxSize)
		for {
			n, err := conn.Read(buff)
			if err != nil {
				stop <- err
				return
			}
			data = append(data, buff[:n]...)
			for {
				offset, serial, buff := protocol.UnPack(data)
				if buff == nil {
					break
				}
				data = data[offset:]
				if f, ok := handles[serial]; ok {
					f(buff)
					break
				}
				log.Printf("protocol(%d) not find\n", serial)
			}
		}
	}(conn, chStop)

	for {
		select {
		case err := <-chStop:
			log.Println(err)
			return
		}
	}
}
