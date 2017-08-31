package main

import (
	"fmt"
	"log"
	"net"

	"github.com/golang/protobuf/proto"
	"github.com/xlplbo/go_protobuf_test/protocol"
)

func senddata(c net.Conn, serial protocol.S2CCmd, msg proto.Message) error {
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
	// Listen on TCP port 2000 on all available unicast and
	// anycast IP addresses of the local system.
	l, err := net.Listen("tcp", ":7788")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		log.Printf("client(%s) connect.\n", conn.RemoteAddr().String())
		go work(conn)
	}
}

func work(conn net.Conn) {
	defer conn.Close()
	var data []byte
	count := 0
	for {
		buf := make([]byte, 1024) //buffer大小应用环境决定
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
			break
		}
		data = append(data, buf[:n]...)
		for {
			if len(data) < 1 {
				break
			}
			offset, serial, buff, err := protocol.UnPack(data)
			if err != nil {
				log.Println(err)
				break
			}
			switch protocol.C2SCmd(serial) {
			case protocol.C2SCmd_None:
				//异常终止
				log.Println("connect close.")
				return
			case protocol.C2SCmd_Login:
				var msgLogin protocol.C2SLogin
				if err := proto.Unmarshal(buff, &msgLogin); err != nil {
					log.Println(err)
				}
				fmt.Printf("login use:%s, passwd:%s\n", msgLogin.GetUser(), msgLogin.GetPasswd())
				// senddata(conn, protocol.S2CCmd_S2C_Result, &protocol.S2CResult{
				// 	Context: "login recviced",
				// })
			case protocol.C2SCmd_Chat:
				var msgChat protocol.C2SChat
				if err := proto.Unmarshal(buff, &msgChat); err != nil {
					log.Println(err)
				}
				fmt.Printf("client say:type %d, %s\n", msgChat.GetCount(), msgChat.GetContext())
				count++
				// senddata(conn, protocol.S2CCmd_S2C_Result, &protocol.S2CResult{
				// 	Context: "chat recviced",
				// })
			}
			data = data[offset:]
		}
	}
	fmt.Printf("count = %d\n", count)
}
