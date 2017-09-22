package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/xlplbo/go_protobuf_test/protocol"
)

var handles map[int32]func([]byte)

func init() {
	handles = make(map[int32]func([]byte))
	handles[int32(protocol.S2CCmd_Result)] = showMsg
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
	chStop := make(chan error, 1)
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
		for {
			buff, err := ioutil.ReadAll(conn)
			if err != nil {
				stop <- err
				return
			}
			data = append(data, buff...)
			for {
				if len(data) < 1 {
					break
				}
				offset, serial, buff, err := protocol.UnPack(data)
				if err != nil {
					stop <- err
					return
				}
				f, ok := handles[serial]
				if !ok {
					log.Printf("protocol(%d) not find\n", serial)
					data = data[offset:]
					continue
				}
				f(buff)
				data = data[offset:]
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
