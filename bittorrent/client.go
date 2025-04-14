package bittorrent

import (
	"fmt"
	"net"
	"time"
)

type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield []byte
	peer     peer
	infoHash [20]byte
	peerID   [20]byte
}

func NewClient(peer peer, infoHash, peerID [20]byte) (*Client, error) {
	target := net.JoinHostPort(peer.IP.String(), fmt.Sprintf("%d", peer.Port))
	conn, err := net.DialTimeout("tcp", target, 3*time.Second)
	if err != nil {
		return nil, err
	}

	err = handshake(conn, infoHash, peerID)
	if err != nil {
		conn.Close()
		return nil, err
	}

	bf, err := recvBitfield(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{Conn: conn, Choked: true, Bitfield: bf, peer: peer, infoHash: infoHash, peerID: peerID}, nil
}
func (c *Client) Read() (*Message, error) {
	msg, err := messageRead(c.Conn)
	return msg, err
}

// SendRequest sends a Request message to the peer
func (c *Client) SendRequest(index, begin, length int) error {
	req := FormatRequest(index, begin, length)
	_, err := c.Conn.Write(req.Serialize())
	return err
}

// SendInterested sends an Interested message to the peer
func (c *Client) SendInterested() error {
	msg := Message{ID: MsgInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendNotInterested sends a NotInterested message to the peer
func (c *Client) SendNotInterested() error {
	msg := Message{ID: MsgNotInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendUnchoke sends an Unchoke message to the peer
func (c *Client) SendUnchoke() error {
	msg := Message{ID: MsgUnchoke}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendHave sends a Have message to the peer
func (c *Client) SendHave(index int) error {
	msg := FormatHave(index)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (bf *Client) HasPiece(index int) bool {
	byteIndex := index / 8
	offset := index % 8
	if byteIndex < 0 || byteIndex >= len(bf.Bitfield) {
		return false
	}
	return bf.Bitfield[byteIndex]>>uint(7-offset)&1 != 0
}

// SetPiece sets a bit in the bitfield
func (bf *Client) SetPiece(index int) {
	byteIndex := index / 8
	offset := index % 8

	// silently discard invalid bounded index
	if byteIndex < 0 || byteIndex >= len(bf.Bitfield) {
		return
	}
	bf.Bitfield[byteIndex] |= 1 << uint(7-offset)
}
