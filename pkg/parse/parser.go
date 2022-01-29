package parse

import (
	"errors"
	"fmt"
	"github.com/ddliu/go-httpclient"
	"github.com/lbbniu/aliyun-m3u8-downloader/pkg/tool"
	"net/url"
)

type Result struct {
	URL  *url.URL
	M3u8 *M3u8
}

func FromURL(link string, aliKey string) (*Result, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	link = u.String()
	resp, err := httpclient.Get(link)
	if err != nil {
		return nil, fmt.Errorf("request m3u8 URL failed: %s", err.Error())
	}
	//noinspection GoUnhandledErrorResult
	m3u8, err := parse(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(m3u8.MasterPlaylist) != 0 {
		sf := m3u8.MasterPlaylist[0]
		return FromURL(tool.ResolveURL(u, sf.URI), aliKey)
	}
	if len(m3u8.Segments) == 0 {
		return nil, errors.New("can not found any TS file description")
	}
	result := &Result{
		URL:  u,
		M3u8: m3u8,
	}

	for idx, key := range m3u8.Keys {
		switch {
		case key.Method == "" || key.Method == CryptMethodNONE:
			continue
		case key.AliyunVoDEncryption && key.Method == CryptMethodAES:
			m3u8.Keys[idx].Key = aliKey
		case !key.AliyunVoDEncryption && key.Method == CryptMethodAES:
			// Request URL to extract decryption key
			keyURL := key.URI
			keyURL = tool.ResolveURL(u, keyURL)
			resp, err := httpclient.Get(keyURL)
			if err != nil {
				return nil, fmt.Errorf("extract key failed: %s", err.Error())
			}
			keyStr, err := resp.ToString()
			if err != nil {
				return nil, err
			}
			fmt.Println("decryption key: ", keyStr)
			m3u8.Keys[idx].Key = keyStr
		default:
			return nil, fmt.Errorf("unknown or unsupported cryption method: %s", key.Method)
		}
	}
	return result, nil
}
