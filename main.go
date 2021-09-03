package main

import (
	"encoding/json"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"speedtest/proxy"
	"speedtest/speedtest"
	"speedtest/utils/retry"
)

type SpeedInfo struct {
	IP       string  `json:"ip"`
	Download float64 `json:"download"`
	Upload   float64 `json:"upload"`
	Ping     int64   `json:"ping"`
	Server   struct {
		Latency int64 `json:"latency"`
	} `json:"server"`
}

var (
	logBool        = flag.Bool("log", false, "print log")
	socks5ProxyStr = flag.String("proxy", "", "socks5 addr")
)

func init() {
	flag.Parse()
	if !*logBool {
		log.SetLevel(log.FatalLevel)
	}
}

func main() {
	var speedInfo SpeedInfo

	if *socks5ProxyStr != "" {
		_, _, err := net.SplitHostPort(*socks5ProxyStr)
		if err != nil {
			fmt.Println("proxy is invalid")
			return
		}

		socks5, _ := proxy.NewProxy(*socks5ProxyStr)
		speedtest.SetSpeedTestSocks5(socks5)
	}

	// find user
	if err := retry.Timed(3, 500).On(func() error {
		user, err := speedtest.FetchUserInfo()
		if err == nil {
			speedInfo.IP = user.IP
		}

		return err
	}); err != nil {
		fmt.Println(err)
		return
	}

	// select best server
	var server speedtest.Server
	if err := retry.Timed(3, 500).On(func() error {
		s, err := speedtest.GetBestServer()
		server = s
		return err
	}); err != nil {
		return
	}

	// test ping
	if err := retry.Timed(1, 500).On(func() error {
		ping, err := speedtest.PingTest(&server)
		speedInfo.Ping = ping
		speedInfo.Server.Latency = ping
		return err
	}); err != nil {
		return
	}

	//
	download, _ := speedtest.DownloadTest(&server)
	upload, _ := speedtest.UploadTest(&server)
	speedInfo.Upload = upload
	speedInfo.Download = download

	data, _ := json.Marshal(&speedInfo)
	fmt.Println(string(data))
}
