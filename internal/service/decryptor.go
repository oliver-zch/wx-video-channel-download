package service

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ChannelsVideoDecryptor 视频解密代理
type ChannelsVideoDecryptor struct {
	client *http.Client
}

func NewChannelsVideoDecryptor() *ChannelsVideoDecryptor {
	tr := &http.Transport{
		TLSNextProto:        make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}
	return &ChannelsVideoDecryptor{
		client: &http.Client{Transport: tr},
	}
}

// DecryptOnly 下载模式：流式解密，带 Content-Disposition: attachment
func (mp *ChannelsVideoDecryptor) DecryptOnly(w http.ResponseWriter, r *http.Request, targetURL string, key uint64, encLimit uint64, filename string) {
	req, err := mp.prepareRequest(r.Method, targetURL, r.Header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	resp, err := mp.client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	var startOffset uint64 = 0
	if cr := resp.Header.Get("Content-Range"); cr != "" {
		parts := strings.Split(cr, " ")
		if len(parts) == 2 {
			rangePart := parts[1]
			dash := strings.Index(rangePart, "-")
			if dash > 0 {
				if v, err := strconv.ParseUint(rangePart[:dash], 10, 64); err == nil {
					startOffset = v
				}
			}
		}
	}
	decryptReader := NewDecryptReader(resp.Body, key, startOffset, encLimit)

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	if w.Header().Get("Accept-Ranges") == "" {
		w.Header().Set("Accept-Ranges", "bytes")
	}

	w.WriteHeader(resp.StatusCode)
	if r.Method == http.MethodHead {
		return
	}
	io.Copy(w, decryptReader)
}

// DecryptOnlyInline 播放模式：流式解密，无 attachment header
func (mp *ChannelsVideoDecryptor) DecryptOnlyInline(w http.ResponseWriter, r *http.Request, targetURL string, key uint64, encLimit uint64) {
	req, err := mp.prepareRequest(r.Method, targetURL, r.Header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	resp, err := mp.client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	var startOffset uint64 = 0
	if cr := resp.Header.Get("Content-Range"); cr != "" {
		parts := strings.Split(cr, " ")
		if len(parts) == 2 {
			rangePart := parts[1]
			dash := strings.Index(rangePart, "-")
			if dash > 0 {
				if v, err := strconv.ParseUint(rangePart[:dash], 10, 64); err == nil {
					startOffset = v
				}
			}
		}
	}
	decryptReader := NewDecryptReader(resp.Body, key, startOffset, encLimit)
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	if w.Header().Get("Accept-Ranges") == "" {
		w.Header().Set("Accept-Ranges", "bytes")
	}
	w.WriteHeader(resp.StatusCode)
	if r.Method == http.MethodHead {
		return
	}
	io.Copy(w, decryptReader)
}

// SimpleProxy 简单代理，不解密
func (mp *ChannelsVideoDecryptor) SimpleProxy(targetURL string, w http.ResponseWriter, r *http.Request) {
	var header http.Header
	method := http.MethodGet
	if r != nil {
		header = r.Header
		method = r.Method
	}
	req, err := mp.prepareRequest(method, targetURL, header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	resp, err := mp.client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	if w.Header().Get("Accept-Ranges") == "" {
		w.Header().Set("Accept-Ranges", "bytes")
	}
	w.WriteHeader(resp.StatusCode)
	if method == http.MethodHead {
		return
	}
	io.Copy(w, resp.Body)
}

func (mp *ChannelsVideoDecryptor) prepareRequest(method, targetURL string, header http.Header) (*http.Request, error) {
	if method != http.MethodGet && method != http.MethodHead {
		method = http.MethodGet
	}
	req, err := http.NewRequest(method, targetURL, nil)
	if err != nil {
		return nil, err
	}
	// 复制 Range header
	if header != nil {
		if rangeHeader := header.Get("Range"); rangeHeader != "" {
			req.Header.Set("Range", rangeHeader)
		}
	}
	return req, nil
}
