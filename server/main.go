package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/xlplbo/go_protobuf_test/protocol"
)

//Player struct
type Player struct {
	index  uint64
	conn   net.Conn
	s      *Server
	chStop chan error
}

//Play Run
func (p *Player) Play() {
	go func() {
		var data []byte
		buff := make([]byte, protocol.MaxSize)
		for {
			n, err := p.conn.Read(buff)
			if err != nil {
				log.Println(err)
				p.Stop()
				return
			}
			data = append(data, buff[:n]...)
			for {
				offset, serial, buff := protocol.UnPack(data)
				if buff == nil {
					break
				}
				data = data[offset:]
				if f, ok := p.s.handles[serial]; ok {
					f(p, buff)
					continue
				}
				log.Printf("protocol id(%d) not handle\n", serial)
			}
		}
	}()
	err := <-p.chStop
	p.conn.Close()
	p.s.DelPlayer(p.index)
	log.Println(err)
}

//Stop player
func (p *Player) Stop() {
	p.chStop <- fmt.Errorf("player(%d) stop", p.index)
}

//GetTargetPlayer ...
func (p *Player) GetTargetPlayer(index uint64) *Player {
	player, ok := p.s.GetPlayer(index)
	if !ok {
		log.Printf("player(%d) non-exsit", index)
		return nil
	}
	return player
}

//SendChat ...
func (p *Player) SendChat(msg string) {
	if err := protocol.Send2Client(p.conn, protocol.S2CCmd_Result, &protocol.S2CResult{
		Context: msg,
	}); err != nil {
		log.Println(err)
	}
}

//GetIndex ...
func (p *Player) GetIndex() uint64 {
	return p.index
}

//Server center
type Server struct {
	index   uint64
	players map[uint64]*Player
	mutex   *sync.RWMutex
	handles map[int32]func(*Player, []byte)
	chStop  chan error
	chConn  chan net.Conn
	chSig   chan os.Signal
}

func (s *Server) getFreeIndex() uint64 {
	var i uint64 = 1
	for i = 1; i <= s.index; i++ {
		if _, ok := s.GetPlayer(i); !ok {
			return i
		}
	}
	s.index++
	return s.index
}

func (s *Server) getPlayerList() []*Player {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	var list []*Player
	for _, p := range s.players {
		list = append(list, p)
	}
	return list
}

//GetPlayer ...
func (s *Server) GetPlayer(index uint64) (*Player, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	player, ok := s.players[index]
	return player, ok
}

func (s *Server) setPlayer(index uint64, p *Player) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.players[index] = p
}

func (s *Server) brocastPlayerList() {
	var buf bytes.Buffer
	buf.WriteString("playerlist:")
	var array []string
	for _, p := range s.getPlayerList() {
		array = append(array, strconv.FormatUint(p.GetIndex(), 10))
	}
	buf.WriteString(strings.Join(array, ","))
	for _, p := range s.getPlayerList() {
		p.SendChat(buf.String() + fmt.Sprintf(" your id: %d", p.GetIndex()))
	}
}

//Run start service
func (s *Server) Run() {
	go func() {
		for {
			conn := <-s.chConn
			index := s.getFreeIndex()
			player := &Player{
				index:  index,
				conn:   conn,
				s:      s,
				chStop: make(chan error),
			}
			s.setPlayer(index, player)
			go player.Play()
			s.brocastPlayerList()
			log.Printf("player(%d) %s connect.\n", index, conn.RemoteAddr().String())
		}
	}()

	msg := <-s.chStop
	for _, p := range s.getPlayerList() {
		p.Stop()
	}
	log.Printf("server stop: %s\n", msg.Error())
}

//ListenTCP only call func use go routine
func (s *Server) ListenTCP(laddr string) {
	l, err := net.Listen("tcp", laddr)
	if err != nil {
		s.chStop <- err
	}
	defer l.Close()
	log.Printf("listen at %s\n", l.Addr().String())
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		s.chConn <- conn
	}
}

//DelPlayer ...
func (s *Server) DelPlayer(key uint64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.players, key)
}

//RegisterHandle ...
func (s *Server) RegisterHandle(id protocol.C2SCmd, f func(*Player, []byte)) {
	nID := int32(id)
	if _, ok := s.handles[nID]; ok {
		log.Printf("protocol(%d) handle repeat\n", nID)
		return
	}
	s.handles[nID] = f
	log.Printf("register handle protocol(%d)\n", nID)
}

//HandleSignal ...
func (s *Server) HandleSignal() {
	signal.Notify(s.chSig, os.Interrupt)
	sig := <-s.chSig
	s.chStop <- fmt.Errorf(sig.String())
}

//NewServer instance
func NewServer() *Server {
	s := &Server{
		index:   0,
		players: make(map[uint64]*Player),
		handles: make(map[int32]func(*Player, []byte)),
		chStop:  make(chan error),
		chConn:  make(chan net.Conn),
		chSig:   make(chan os.Signal),
		mutex:   &sync.RWMutex{},
	}
	s.RegisterHandle(protocol.C2SCmd_Abnormal, func(p *Player, msg []byte) {
		p.Stop()
	})
	s.RegisterHandle(protocol.C2SCmd_Chat, func(p *Player, msg []byte) {
		var chatMsg protocol.C2SChat
		if err := proto.Unmarshal(msg, &chatMsg); err != nil {
			p.Stop()
		}
		player := p.GetTargetPlayer(chatMsg.Index)
		if player != nil {
			player.SendChat(chatMsg.Context)
		}
	})
	return s
}

func main() {
	app := NewServer()
	go app.HandleSignal()
	go app.ListenTCP(":7788")
	app.Run()
}
