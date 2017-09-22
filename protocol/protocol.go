package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"unsafe"

	"github.com/golang/protobuf/proto"
)

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

// Pack data params serial:protocol id, m: protocol message
// package = head(4 byte) + body proto(* byte)
// return []byte buff, error
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
	//fmt.Printf("pack %d\n", len(body))
	head, err := int2bytes(len(body))
	if err != nil {
		return nil, err
	}
	return append(head, body...), nil
}

// UnPack data params data buff []byte
// return data buff offset, protocol id, protocol message buff []byte, error
func UnPack(data []byte) (int, int32, []byte, error) {
	if len(data) < headsize {
		return 0, 0, nil, fmt.Errorf("data length < %d", headsize)
	}
	size, err := bytes2int(data[:headsize])
	//fmt.Printf("unpack %d\n", size)
	if err != nil {
		return 0, 0, nil, err
	}
	size += headsize
	if len(data) < size {
		return 0, 0, nil, fmt.Errorf("data length < %d", size)
	}
	var pkg Package
	if err := proto.Unmarshal(data[headsize:size], &pkg); err != nil {
		//data abnormal
		return 0, 0, nil, nil
	}
	return size, pkg.GetSerial(), pkg.GetBuff(), nil
}

// SendMessage protocol id, protocol message
// return error
func SendMessage(conn net.Conn, serial int32, msg proto.Message) error {
	buff, err := Pack(serial, msg)
	if err != nil {
		return err
	}
	if _, err := conn.Write(buff); err != nil {
		return err
	}
	return nil
}
