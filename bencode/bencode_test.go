package bencode

import (
	"bufio"
	"os"
	"testing"
)

func TestEncode(t *testing.T) {
	file, err := os.Open("./debian.torrent")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	bufioReader := bufio.NewReader(file)
	data, err := Decode(bufioReader)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(data.(map[string]interface{})["url-list"])
}
