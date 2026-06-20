package webrtc

import (
	"encoding/binary"
	"log/slog"
	"net"
)

const (
	stunBindingRequest  = 0x0001
	stunBindingResponse = 0x0101
	stunXorMappedAddr   = 0x0020
)

var stunMagicCookie = [4]byte{0x21, 0x12, 0xA4, 0x42}

func StartSTUNServer(addr string) error {
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	slog.Info("STUN server listening", "addr", addr)

	buf := make([]byte, 1500)
	for {
		n, remote, err := conn.ReadFrom(buf)
		if err != nil {
			continue
		}
		if n < 20 {
			continue
		}

		msgType := binary.BigEndian.Uint16(buf[0:2])
		if msgType != stunBindingRequest {
			continue
		}

		magicCookie := binary.BigEndian.Uint32(buf[4:8])
		if magicCookie != 0x2112A442 {
			continue
		}

		txnID := make([]byte, 12)
		copy(txnID, buf[8:20])

		udpAddr, ok := remote.(*net.UDPAddr)
		if !ok {
			continue
		}

		attrs := buildXORMappedAddress(udpAddr.IP.To4(), uint16(udpAddr.Port), txnID)

		resp := make([]byte, 20+len(attrs))
		binary.BigEndian.PutUint16(resp[0:2], stunBindingResponse)
		binary.BigEndian.PutUint16(resp[2:4], uint16(len(attrs)))
		binary.BigEndian.PutUint32(resp[4:8], 0x2112A442)
		copy(resp[8:20], txnID)
		copy(resp[20:], attrs)

		conn.WriteTo(resp, remote)
	}
}

func buildXORMappedAddress(ip net.IP, port uint16, txnID []byte) []byte {
	attr := make([]byte, 12)
	binary.BigEndian.PutUint16(attr[0:2], stunXorMappedAddr)
	binary.BigEndian.PutUint16(attr[2:4], 8)
	attr[4] = 0x01

	xorPort := port ^ 0xa442
	binary.BigEndian.PutUint16(attr[8:10], xorPort)

	for i := 0; i < 4; i++ {
		attr[5+i] = ip[i] ^ stunMagicCookie[i]
	}

	return attr
}
