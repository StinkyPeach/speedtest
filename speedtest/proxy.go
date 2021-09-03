package speedtest

import (
	socks5 "speedtest/proxy"
)

var (
	speedProxy *socks5.Socks5Proxy
)

func GetSpeedTestSocks5() *socks5.Socks5Proxy {
	return speedProxy
}

func SetSpeedTestSocks5(proxy *socks5.Socks5Proxy) {
	speedProxy = proxy
}
