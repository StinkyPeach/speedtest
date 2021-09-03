package proxy

import (
	"io"
	"net/http"
	"testing"
)

func TestDialContext(t *testing.T) {
	proxy, _ := NewProxy("192.168.8.49:1080")
	dial := proxy.DialContext()

	req, _ := http.NewRequest(http.MethodGet, "https://ifconfig.me", nil)
	client := http.Client{
		Transport: &http.Transport{
			DialContext: dial,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}

	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	t.Log(string(data))
}

func TestProxy(t *testing.T) {
	proxy, _ := NewProxy("192.168.8.49:1080")
	dial := proxy.Proxy()

	req, _ := http.NewRequest(http.MethodGet, "https://ifconfig.me", nil)
	client := http.Client{
		Transport: &http.Transport{
			Proxy: dial,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}

	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	t.Log(string(data))
}