package protocol

import (
	"bytes"
	"encoding/binary"
	"net"
	"unsafe"

	"github.com/golang/protobuf/proto"
)

//MaxSize conn.Read max buffer size
//Set the size according to the actual application environment
const MaxSize = 1024

const headsize = int(unsafe.Sizeof(uint32(0)))

func int2bytes(value int) ([]byte, error) {
	pbuffer := bytes.NewBuffer([]byte{})
	if err := binary.Write(pbuffer, binary.BigEndian, uint32(value)); err != nil {
		return nil, err
	}
	return pbuffer.Bytes()[:headsize], nil
}

func bytes2int(b []byte) (int, error) {
	var value uint32
	pbuffer := bytes.NewBuffer(b[:headsize])
	if err := binary.Read(pbuffer, binary.BigEndian, &value); err != nil {
		return 0, err
	}
	return int(value), nil
}

// Pack params protocol id, *proto.Message
// package = head(4 byte) + body([]byte)
// return []byte, error
func Pack(serial int32, m proto.Message) ([]byte, error) {
	var pkg Package
	buff, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	pkg.Serial = serial
	pkg.Buff = buff
	body, err := proto.Marshal(&pkg)
	if err != nil {
		return nil, err
	}
	head, err := int2bytes(len(body))
	if err != nil {
		return nil, err
	}
	return append(head, body...), nil
}

// UnPack params []byte
// return offset, protocol id,  body([]byte)
func UnPack(data []byte) (int, int32, []byte) {
	if len(data) < headsize {
		return 0, 0, nil
	}
	bodysize, err := bytes2int(data[:headsize])
	if err != nil {
		return 0, 0, nil
	}
	offset := headsize + bodysize
	if len(data) < offset {
		return 0, 0, nil
	}
	var pkg Package
	if err := proto.Unmarshal(data[headsize:offset], &pkg); err != nil {
		return 0, 0, []byte{} //data abnormal and disconnect
	}
	return offset, pkg.Serial, pkg.Buff
}

// Send2Client protocol id, protocol message
// return error
func Send2Client(conn net.Conn, serial S2CCmd, msg proto.Message) error {
	buff, err := Pack(int32(serial), msg)
	if err != nil {
		return err
	}
	if _, err := conn.Write(buff); err != nil {
		return err
	}
	return nil
}

// Send2Server protocol id, protocol message
// return error
func Send2Server(conn net.Conn, serial C2SCmd, msg proto.Message) error {
	buff, err := Pack(int32(serial), msg)
	if err != nil {
		return err
	}
	if _, err := conn.Write(buff); err != nil {
		return err
	}
	return nil
}
