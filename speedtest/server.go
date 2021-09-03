package speedtest

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

const speedTestApi = "https://www.speedtest.net/api/js/servers?engine=js"

// Server is the server struct for a json server
type Server struct {
	URL             string `json:"url"`
	Lat             string `json:"lat"`
	Lon             string `json:"lon"`
	Distance        int    `json:"distance"`
	Name            string `json:"name"`
	Country         string `json:"country"`
	Cc              string `json:"cc"`
	Sponsor         string `json:"sponsor"`
	ID              string `json:"id"`
	Preferred       int    `json:"preferred"`
	Host            string `json:"host"`
	ForcePingSelect int    `json:"force_ping_select"`
	BestTestPing    int64
}

// GetAllServers parses the list of all recommended servers
func GetAllServers() (servers []Server, err error) {
	client := http.Client{}
	client.Timeout = time.Second * 30
	socks5 := GetSpeedTestSocks5()
	if socks5 != nil {
		client.Transport = &http.Transport{
			DialContext: socks5.DialContext(),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	req, _ := http.NewRequest(http.MethodGet, speedTestApi, nil)
	resp, err := client.Do(req)
	if err != nil {
		return servers, err
	}

	defer client.CloseIdleConnections()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return servers, err
	}

	err = json.Unmarshal(body, &servers)
	if err != nil {
		return servers, err
	}

	return servers, err
}
