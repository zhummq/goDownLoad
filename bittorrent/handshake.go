package bittorrent

import (
	"fmt"
	"io"
	"net"
	"time"
)

func handshake(conn net.Conn, InfoHash, PeerID [20]byte) error {

	conn.SetDeadline(time.Now().Add(20 * time.Second))
	defer conn.SetDeadline(time.Time{})
	handMessage := Serialize("BitTorrent protocol", InfoHash, PeerID)
	_, err := conn.Write(handMessage)
	if err != nil {
		return err
	}
	err = ReadisOK(conn, InfoHash)

	if err != nil {
		//log.Printf("handshake failed: %s", err)
		return err
	}

	return nil
}

func Serialize(Pstr string, InfoHash, PeerID [20]byte) []byte {
	buf := make([]byte, len(Pstr)+49)
	buf[0] = byte(len(Pstr))
	curr := 1
	curr += copy(buf[curr:], Pstr)
	curr += copy(buf[curr:], make([]byte, 8))
	curr += copy(buf[curr:], InfoHash[:])
	curr += copy(buf[curr:], PeerID[:])
	return buf
}

func recvBitfield(conn net.Conn) ([]byte, error) {
	conn.SetDeadline(time.Now().Add(20 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	msg, err := messageRead(conn)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		err := fmt.Errorf("expected bitfield but got ")
		return nil, err
	}
	if msg.ID != MsgBitfield {
		err := fmt.Errorf("expected bitfield but got id %d", msg.ID)
		return nil, err
	}

	return msg.Payload, nil
}

func ReadisOK(r io.Reader, info [20]byte) error {
	lengthBuf := make([]byte, 1)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		if err == io.EOF {
			return fmt.Errorf("unexpected EOF")
		}
		return err
	}
	pstrlen := int(lengthBuf[0])

	if pstrlen == 0 {
		err := fmt.Errorf("pstrlen cannot be 0")
		return err
	}

	handshakeBuf := make([]byte, 48+pstrlen)
	_, err = io.ReadFull(r, handshakeBuf)
	if err != nil {
		return err
	}

	var infoHash, peerID [20]byte

	copy(infoHash[:], handshakeBuf[pstrlen+8:pstrlen+8+20])
	copy(peerID[:], handshakeBuf[pstrlen+8+20:])
	if infoHash != info {
		err := fmt.Errorf("infoHash mismatch")
		return err
	}
	return nil
}
