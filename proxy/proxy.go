package proxy

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"golang.org/x/net/proxy"
)

type Socks5Proxy struct {
	Host string
	Port uint
}

func NewProxy(address string) (*Socks5Proxy, error) {
	host, p, err := net.SplitHostPort(address)
	if err != nil {
		return nil, errors.New("proxy address is valid")
	}

	port, err := strconv.Atoi(p)
	if err != nil {
		return nil, errors.New("proxy address is valid")
	}

	return &Socks5Proxy{
		host,
		uint(port),
	}, nil
}

func (s *Socks5Proxy) String() string {
	return net.JoinHostPort(s.Host, strconv.Itoa(int(s.Port)))
}

func (s *Socks5Proxy) Socks5Url() string {
	return "socks5://" + s.String()
}

func (s *Socks5Proxy) DialContext() func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		socks5, err := proxy.SOCKS5("tcp", s.String(), nil, proxy.Direct)
		if err != nil {
			return nil, err
		}

		return socks5.Dial(network, addr)
	}
}

func (s *Socks5Proxy) Proxy() func(r *http.Request) (*url.URL, error) {
	return func(r *http.Request) (*url.URL, error) {
		proxy, err := url.Parse(s.Socks5Url())
		if err != nil {
			return nil, err
		}

		return proxy, nil
	}
}
