package bittorrent

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"log"
	"unsafe"

	"github.com/schollz/progressbar/v3"
	"github.com/zhummq/bencode"
	"golang.org/x/sys/windows"
)

type torrent struct {
	Announce    string
	InfoHash    [20]byte
	PeerID      [20]byte
	Port        uint16
	pieces      [][20]byte
	Length      int
	Internal    int
	Peers       []peer
	PieceLength int
	Outdata     []byte
}

func NewTorrent(data interface{}) (*torrent, error) {
	peerID := make([]byte, 20)
	_, err := rand.Read(peerID)
	if err != nil {
		return nil, err
	}

	infoHashString := data.(map[string]interface{})["info"].(map[string]interface{})
	infoBytes, err := bencode.Encode(infoHashString)
	if err != nil {
		return nil, err
	}
	infoHash := sha1.Sum(infoBytes)

	piecesArray := infoHashString["pieces"].(string)
	pieces, err := splitPieceHashes(piecesArray)
	if err != nil {
		return nil, err
	}
	return &torrent{
		Announce:    data.(map[string]interface{})["announce"].(string),
		InfoHash:    infoHash,
		PeerID:      [20]byte(peerID),
		Port:        6881,
		Length:      int(data.(map[string]interface{})["info"].(map[string]interface{})["length"].(int64)),
		pieces:      pieces,
		PieceLength: int(infoHashString["piece length"].(int64)),
	}, nil
}

func splitPieceHashes(pieces string) ([][20]byte, error) {
	hashLen := 20
	buf := []byte(pieces)
	if len(buf)%hashLen != 0 {
		return nil, fmt.Errorf("received malformed pieces of length %d", len(buf))
	}
	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}
	return hashes, nil
}

func (t *torrent) startDownloadWorker(peer peer, workQueue chan *pieceWork, results chan *pieceResult) {
	c, err := NewClient(peer, t.InfoHash, t.PeerID)
	if err != nil {
		//log.Printf("Could not handshake with %v %v %v. Disconnecting\n", err, peer.IP, peer.Port)
		return
	}

	defer c.Conn.Close()
	//log.Printf("connected to %s\n", peer.IP)

	c.SendUnchoke()
	c.SendInterested()

	for pw := range workQueue {
		if !c.HasPiece(pw.index) {
			workQueue <- pw // Put piece back on the queue
			continue
		}

		// Download the piece
		buf, err := attemptDownloadPiece(c, pw)
		if err != nil {
			log.Println("Exiting", err)
			workQueue <- pw // Put piece back on the queue
			return
		}

		err = checkIntegrity(pw, buf)
		if err != nil {
			log.Printf("Piece #%d failed integrity check\n", pw.index)
			workQueue <- pw // Put piece back on the queue
			continue
		}

		c.SendHave(pw.index)
		begin, end := t.calculateBoundsForPiece(pw.index)
		copy(t.Outdata[begin:end], buf)
		windows.FlushViewOfFile(uintptr(unsafe.Pointer(&t.Outdata[begin])), uintptr(end-begin))
		results <- &pieceResult{pw.index, (end - begin)}
	}
}

func (t *torrent) calculateBoundsForPiece(index int) (begin int, end int) {
	begin = index * t.PieceLength
	end = begin + t.PieceLength
	if end > int(t.Length) {
		end = t.Length
	}
	return begin, end
}

func (t *torrent) calculatePieceSize(index int) int {
	begin, end := t.calculateBoundsForPiece(index)
	return end - begin
}

// Download downloads the torrent. This stores the entire file in memory.

func (t *torrent) Download() error {
	workQueue := make(chan *pieceWork, len(t.pieces))
	results := make(chan *pieceResult)
	for index, hash := range t.pieces {
		length := t.calculatePieceSize(index)
		workQueue <- &pieceWork{index, hash, length}
	}

	// Start workers
	for _, peer := range t.Peers {
		go t.startDownloadWorker(peer, workQueue, results)
	}
	// Collect results into a buffer until full
	pb := progressbar.NewOptions(
		t.Length,
		progressbar.OptionSetWriter(log.Writer()),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetDescription("Downloading"),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionEnableColorCodes(true),
		//progressbar.OptionThrottle(65*time.Millisecond),
	)
	donePieces := 0
	for donePieces < len(t.pieces) {
		res := <-results
		donePieces++

		pb.Add(res.length)
	}
	pb.Finish()
	close(workQueue)
	return nil
}
