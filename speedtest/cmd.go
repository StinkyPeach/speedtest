package speedtest

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net"
	"time"
)

type Result struct {
	Start      time.Time
	Finish     time.Time
	DurationMs int64
	Bytes      int
}

// Dial will return a socket connection from the server
func Dial(server string) (conn net.Conn, err error) {
	return connect(server)
}

// Ping issues and times a ping command
func ping(conn net.Conn) (ms int64) {
	start := time.Now()
	cmdString := fmt.Sprintf("PING %s", start)
	command(conn, cmdString)
	finish := time.Now()
	ms = calcMs(start, finish)
	log.WithFields(log.Fields{
		"ms": ms,
		"package": GetPackagePath(PathObjHolder{}),
		"function": "ping",
	}).Debug("ping")
	return ms
}

// download performs a timed download, returning the Mbps
func download(conn net.Conn, numBytes int) (result Result) {
	start := time.Now()
	cmdString := fmt.Sprintf("DOWNLOAD %d", numBytes)
	send(conn, cmdString)
	data, err := receive(conn)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"package": GetPackagePath(PathObjHolder{}),
			"function": "download",
		}).Debug()
		return Result{
			Bytes: len(data),
		}
	}

	finish := time.Now()
	result = Result{
		Start:      start,
		Finish:     finish,
		DurationMs: calcMs(start, finish),
		Bytes:      numBytes,
	}

	log.WithFields(log.Fields{
		"duration": result.DurationMs,
		"bytes":    result.Bytes,
		"package": GetPackagePath(PathObjHolder{}),
		"function": "download",
	}).Debug("download")

	return result
}

// upload performs a timed upload of numBytes random bytes, returning the mpbs
func upload(conn net.Conn, numBytes int) (result Result) {
	randBytes := generateBytes(numBytes)

	bytesString := fmt.Sprintf("%d", len(randBytes))
	lenBytesString := len(bytesString)
	finalBytes := lenBytesString + numBytes + len("UPLOAD_0_\n\n")

	cmdString1 := fmt.Sprintf("UPLOAD %d 0", finalBytes)
	cmdString2 := fmt.Sprintf("%s", randBytes)

	start := time.Now()
	send(conn, cmdString1)
	send(conn, cmdString2)
	_, err := receive(conn)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"package": GetPackagePath(PathObjHolder{}),
			"function": "upload",
		}).Info()

		return Result{
			Bytes: 0,
		}
	}

	finish := time.Now()
	result = Result{
		Start:      start,
		Finish:     finish,
		DurationMs: calcMs(start, finish),
		Bytes:      numBytes,
	}

	log.WithFields(log.Fields{
		"duration": result.DurationMs,
		"bytes":    result.Bytes,
		"package": GetPackagePath(PathObjHolder{}),
		"function": "upload",
	}).Debug("upload")

	return result
}

func generateBytes(numBytes int) (random []byte) {
	log.WithFields(log.Fields{
		"numBytes": numBytes,
		"package": GetPackagePath(PathObjHolder{}),
		"function": "generateBytes",
	}).Debug("generateBytes")
	random = make([]byte, numBytes)
	_, err := rand.Read(random)
	if err != nil {
		random = generateBytes(numBytes)
	}
	return random
}

func calcMs(start time.Time, finish time.Time) (ms int64) {
	diff := finish.Sub(start)
	return diff.Nanoseconds() / 1000000
}
