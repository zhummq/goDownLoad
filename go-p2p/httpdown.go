package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"unsafe"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/sys/windows"
)

type task struct {
	start, end int
}

type result struct {
	length int
}

const (
	MaxBlockSize = 1024 * 256
	Number       = 16
)

func httpdownload(url, outPath string) {
	// download the file
	resp, err := http.Head(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	length := int(resp.ContentLength)
	if resp.Header.Get("Accept-Ranges") != "bytes" {
		fmt.Println("Server does not support range requests")
		return
	}
	n := length / MaxBlockSize
	if length%MaxBlockSize != 0 {
		n++
	}
	tasks := make(chan *task, n)
	for i := range n {
		start := i * MaxBlockSize
		end := min(start+MaxBlockSize, length)
		tasks <- &task{start, end}
	}
	results := make(chan *result)
	outFile, err := os.Create(outPath)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()
	err = outFile.Truncate(int64(length))
	if err != nil {
		log.Fatal(err)
	}
	mmap, err := windows.CreateFileMapping(windows.Handle(outFile.Fd()), nil, windows.PAGE_READWRITE, 0, uint32(length), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer windows.CloseHandle(mmap)
	addr, err := windows.MapViewOfFile(mmap, windows.FILE_MAP_WRITE, 0, 0, uintptr(length))
	if err != nil {
		log.Fatal(err)
	}
	defer windows.UnmapViewOfFile(addr)
	outdata := (*[1 << 30]byte)(unsafe.Pointer(addr))[:length]
	for range Number {
		go downloadBlock(url, tasks, results, outdata)
	}
	pb := progressbar.NewOptions(
		length,
		progressbar.OptionSetWriter(log.Writer()),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetDescription("Downloading"),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionEnableColorCodes(true),
	)
	size := 0
	for size < n {
		re := <-results
		size++
		pb.Add(re.length)
	}

	pb.Finish()
	close(tasks)
}

func downloadBlock(url string, tasks chan *task, results chan *result, outdata []byte) {
	for task := range tasks {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			tasks <- task
			log.Printf("%s", err.Error())
			continue
		}
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", task.start, task.end))
		client := http.Client{
			Timeout: 30 * time.Second,
		}
		resp, err := client.Do(req)
		if err != nil {
			tasks <- task
			log.Printf("%s", err.Error())
			continue
		}
		len := task.end - task.start
		buf := make([]byte, len)
		_, err = io.ReadFull(resp.Body, buf)
		if err != nil {
			tasks <- task
			log.Printf("%s", err.Error())
			continue
		}
		copy(outdata[task.start:task.end], buf)
		windows.FlushViewOfFile(uintptr(unsafe.Pointer(&outdata[task.start])), uintptr(len))
		results <- &result{len}
		resp.Body.Close()
	}
}
