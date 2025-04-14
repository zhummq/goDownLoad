package main

import (
	"bufio"
	"log"
	"os"
	"unsafe"

	"github.com/zhummq/bencode"
	"github.com/zhummq/bittorrent"
	"golang.org/x/sys/windows"
)

func download(inPath, outPath string) {
	file, err := os.Open(inPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	bufioReader := bufio.NewReader(file)
	data, err := bencode.Decode(bufioReader)
	if err != nil {
		log.Fatal(err)
	}
	torrent, err := bittorrent.NewTorrent(data)
	if err != nil {
		log.Fatal(err)
	}
	content, err := torrent.GetTracker()
	if err != nil {
		log.Fatal(err)
	}
	torrent.Peers, err = torrent.GetPeers(content.(map[string]interface{})["peers"].(string))
	if err != nil {
		log.Fatal(err)
	}
	torrent.Internal = int(content.(map[string]interface{})["interval"].(int64))

	// use memory-mapped file to write to disk
	outFile, err := os.Create(outPath)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()
	err = outFile.Truncate(int64(torrent.Length))
	if err != nil {
		log.Fatal(err)
	}
	mmap, err := windows.CreateFileMapping(windows.Handle(outFile.Fd()), nil, windows.PAGE_READWRITE, 0, uint32(torrent.Length), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer windows.CloseHandle(mmap)
	addr, err := windows.MapViewOfFile(mmap, windows.FILE_MAP_WRITE, 0, 0, uintptr(torrent.Length))
	if err != nil {
		log.Fatal(err)
	}
	defer windows.UnmapViewOfFile(addr)
	outdata := (*[1 << 30]byte)(unsafe.Pointer(addr))[:torrent.Length]
	torrent.Outdata = outdata
	torrent.Download()
}
