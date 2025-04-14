package bittorrent

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/zhummq/bencode"
)

func (t *torrent) GetTracker() (interface{}, error) {
	url, err := t.getUrl()
	if err != nil {
		return nil, err
	}
	c := &http.Client{Timeout: 15 * time.Second}
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := bencode.Decode(resp.Body)
	if err != nil {
	}
	return data, nil
}

func (t *torrent) getUrl() (string, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}
	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(t.PeerID[:])},
		"port":       []string{strconv.Itoa(int(t.Port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"left":       []string{strconv.Itoa(t.Length)},
		"compact":    []string{"1"},
	}
	base.RawQuery = params.Encode()
	return base.String(), nil
}
