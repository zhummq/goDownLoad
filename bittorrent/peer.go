package bittorrent

import (
	"encoding/binary"
	"net"
)

type peer struct {
	IP   net.IP
	Port uint16
}

func (p *torrent) GetPeers(peersString string) ([]peer, error) {
	const peerSize = 6 // 4 for IP, 2 for port
	numPeers := len(peersString) / peerSize
	if len(peersString)%peerSize != 0 {
		return nil, nil
	}
	peers := make([]peer, numPeers)
	for i := range numPeers {
		offset := i * peerSize
		peers[i].IP = net.IP(peersString[offset : offset+4])
		peers[i].Port = binary.BigEndian.Uint16([]byte(peersString[offset+4 : offset+6]))
	}
	return peers, nil
}
