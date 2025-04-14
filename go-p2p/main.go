package main

import "os"

func main() {

	inPath := os.Args[1]
	outPath := os.Args[2]
	_, err := os.Stat(inPath)
	if err != nil {
		httpdownload(inPath, outPath)
	} else {
		download(inPath, outPath)
	}

}
