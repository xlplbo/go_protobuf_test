package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/xlplbo/go_protobuf_test/protocol"
)

var handles map[int32]func([]byte)
var chStop chan error
var chSig chan os.Signal

func init() {
	chStop = make(chan error)
	chSig = make(chan os.Signal)
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
	log.Printf("register handle protocol(%d)\n", nID)
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

func handleSignal() {
	signal.Notify(chSig, os.Interrupt)
	s := <-chSig
	chStop <- fmt.Errorf("stop clinet: %s", s.String())
}

func main() {
	go handleSignal()

	//Dail TCP
	chConn1 := make(chan net.Conn)
	chConn2 := make(chan net.Conn)
	go func(ch1, ch2 chan<- net.Conn) {
		for {
			select {
			case <-time.Tick(time.Second):
				log.Println("connect server...")
				conn, err := net.Dial("tcp", "127.0.0.1:7788")
				if err != nil {
					log.Println(err)
					continue
				}
				log.Printf("%s established", conn.RemoteAddr().String())
				ch1 <- conn
				ch2 <- conn
				return
			}
		}
	}(chConn1, chConn2)

	// Read data
	go func(ch <-chan net.Conn) {
		conn := <-ch
		var data []byte
		buff := make([]byte, protocol.MaxSize)
		for {
			n, err := conn.Read(buff)
			if err != nil {
				chStop <- err
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
					continue
				}
				log.Printf("protocol(%d) not find\n", serial)
			}
		}
	}(chConn1)

	// Send data
	go func(ch <-chan net.Conn) {
		conn := <-ch
		for {
			var input string
			_, err := fmt.Scanln(&input)
			if err != nil {
				log.Println(err)
				log.Println("please input: target id:msg context")
				continue
			}
			v := strings.Split(input, ":")
			if len(v) != 2 {
				log.Println("please input: target id:msg context")
				continue
			}
			index, err := strconv.ParseUint(v[0], 10, 64)
			if err != nil {
				log.Println(err)
				log.Println("please input: target id:msg context")
				continue
			}
			protocol.Send2Server(conn, protocol.C2SCmd_Chat, &protocol.C2SChat{
				Index:   index,
				Context: v[1],
			})
		}
	}(chConn2)

	// Stress test
	// go func(ch <-chan net.Conn) {
	// 	conn := <-ch
	// 	for {
	// 		protocol.Send2Server(conn, protocol.C2SCmd_Chat, &protocol.C2SChat{
	// 			Index:   1,
	// 			Context: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	// 		})
	// 	}
	// }(chConn2)

	log.Println(<-chStop)
}
