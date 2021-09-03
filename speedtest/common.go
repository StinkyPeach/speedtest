package speedtest

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"reflect"
	"strings"
	"time"
)

// connect starts a tcp connection
func connect(server string) (conn net.Conn, err error) {
	socks5 := GetSpeedTestSocks5()
	if socks5 != nil {
		conn, err = socks5.DialContext()(context.Background(), "tcp", server)
	} else {
		conn, err = net.DialTimeout("tcp", server, time.Second * 10)
	}

	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"package": GetPackagePath(PathObjHolder{}),
			"function": "connect",
		}).Info()
	}
	return conn, err
}

// send the string to connection
func send(conn net.Conn, msg string) (send int64, err error) {
	nm := fmt.Sprintf("%s\n", msg)
	log.WithFields(log.Fields{
		"len": len(nm),
		"package": GetPackagePath(PathObjHolder{}),
		"function": "send",
	}).Debug("Tx")

	send, err = io.Copy(conn, bytes.NewReader([]byte(nm)))
	return send, err
}

// receive gets the next line from the connection
func receive(conn net.Conn) (status []byte, err error) {
	data, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"package": GetPackagePath(PathObjHolder{}),
			"function": "receive",
		}).Debug()
	}

	return data, err
}

// command is a shortcut to send and receive
func command(conn net.Conn, command string) (resp string, err error) {
	_, err = send(conn, command)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"package": GetPackagePath(PathObjHolder{}),
			"function": "command",
		}).Debug()
	}
	data, err := receive(conn)
	resp = strings.TrimSpace(string(data))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"package": GetPackagePath(PathObjHolder{}),
			"function": "command",
		}).Debug()
	}
	return resp, err
}

type PathObjHolder struct{}

func GetPackagePath(obj interface{}) string {
	return reflect.TypeOf(obj).PkgPath()
}
