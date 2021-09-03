package speedtest

import (
	"context"
	log "github.com/sirupsen/logrus"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var thread int = 4

// GetBestServer gets the first in the list
func GetBestServer() (bestsServer Server, err error) {
	var bestSpeed int64 = 999
	servers, err := GetAllServers()
	if err != nil {
		log.WithFields(log.Fields{
			"err":      err,
			"package":  GetPackagePath(PathObjHolder{}),
			"function": "GetBestServer",
		}).Info()
	}

	for s := range servers {
		log.WithFields(log.Fields{
			"server":   servers[s].ID,
			"package":  GetPackagePath(PathObjHolder{}),
			"function": "GetBestServer",
		}).Debug("GetBestServer")

		c, err := connect(servers[s].Host)
		if err != nil {
			continue
		}

		res := Ping(c, 3)
		servers[s].BestTestPing = res
		log.WithFields(log.Fields{
			"speed":    res,
			"package":  GetPackagePath(PathObjHolder{}),
			"function": "GetBestServer",
		}).Debug("GetBestServer")

		if res < bestSpeed {
			bestSpeed = res
			bestsServer = servers[s]
		}

		c.Close()
	}

	log.WithFields(log.Fields{
		"bestsServer": bestsServer,
		"package":     GetPackagePath(PathObjHolder{}),
		"function":    "GetBestServer",
	}).Debug("GetBestServer")

	return bestsServer, err
}

// Upload runs download tests through numBytes until a download test takes more than numSeconds
func Upload(conn net.Conn, seedBytes int, ctx context.Context) (total int) {
	var numTests int = 0

upload:
	for {
		seedBytes = seedBytes + 1024*1024
		select {
		case <-ctx.Done():
			break upload
		default:
			conn.SetDeadline(time.Now().Add(time.Second * 16))
			res := upload(conn, seedBytes)
			conn.SetDeadline(time.Time{})
			total = total + res.Bytes
			numTests++
			log.WithFields(log.Fields{
				"package":  GetPackagePath(PathObjHolder{}),
				"function": "Upload",
				"testNum":  numTests,
			}).Info()
		}

	}

	log.WithFields(log.Fields{
		"total":    total,
		"package":  GetPackagePath(PathObjHolder{}),
		"function": "Upload",
	}).Info("Upload")

	return total
}

func UploadTest(s *Server) (result float64, error error) {
	var totalBytes int64
	var wg sync.WaitGroup

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)

	uploadFun := func(conn net.Conn, seedBytes int, ctx context.Context) {
		bytes := Upload(conn, seedBytes, ctx)
		atomic.AddInt64(&totalBytes, int64(bytes))
		wg.Done()
	}

	connSlice := make([]net.Conn, 0)
	for i := 0; i < thread; i++ {
		conn, err := Dial(s.Host)
		if err != nil {
			log.WithFields(log.Fields{
				"err":      err.Error(),
				"package":  GetPackagePath(PathObjHolder{}),
				"function": "UploadTest",
			}).Info("UploadTest")

			continue
		}

		connSlice = append(connSlice, conn)
		error = err
	}

	sTime := time.Now()
	for _, v := range connSlice {
		wg.Add(1)
		go uploadFun(v, 1<<20, ctx)
	}

	wg.Wait()
	result = calcMbps(sTime, time.Now(), totalBytes)

	for _, v := range connSlice {
		v.Close()
	}

	return result, error
}

// Download runs download tests through numBytes until a download test takes more than numSeconds
func Download(conn net.Conn, seedBytes int, ctx context.Context) (total int) {
	var numTests int = 0

download:
	for {
		seedBytes = seedBytes + 1024*1024
		select {
		case <-ctx.Done():
			break download
		default:
			conn.SetDeadline(time.Now().Add(time.Second * 16))
			res := download(conn, seedBytes)
			conn.SetDeadline(time.Time{})
			total += res.Bytes
			numTests++
			log.WithFields(log.Fields{
				"package":  GetPackagePath(PathObjHolder{}),
				"function": "Download",
				"testNum":  numTests,
			}).Info()
		}
	}

	log.WithFields(log.Fields{
		"total":    total,
		"package":  "app",
		"function": "Download",
	}).Info("Download")

	return total
}

func DownloadTest(s *Server) (result float64, error error) {
	var totalBytes int64
	var wg sync.WaitGroup

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)

	uploadFun := func(conn net.Conn, seedBytes int, ctx context.Context) {
		bytes := Download(conn, seedBytes, ctx)
		atomic.AddInt64(&totalBytes, int64(bytes))
		wg.Done()
	}

	connSlice := make([]net.Conn, 0)
	for i := 0; i < thread; i++ {
		conn, err := Dial(s.Host)
		if err != nil {
			log.WithFields(log.Fields{
				"err":      err.Error(),
				"package":  GetPackagePath(PathObjHolder{}),
				"function": "DownloadTest",
			}).Info("DownloadTest")

			continue
		}
		connSlice = append(connSlice, conn)
		error = err
	}

	sTime := time.Now()
	for _, v := range connSlice {
		wg.Add(1)
		go uploadFun(v, 1<<20, ctx)
	}

	wg.Wait()
	result = calcMbps(sTime, time.Now(), totalBytes)

	for _, v := range connSlice {
		v.Close()
	}

	return result, error
}

// Ping gets roundTrip time to issue the "PING" command
func Ping(conn net.Conn, numTests int) (results int64) {
	var acc int64

	for i := 1; i <= numTests; i++ {
		res := ping(conn)
		acc = acc + res
	}

	results = acc / int64(numTests)
	log.WithFields(log.Fields{
		"results":  results,
		"package":  GetPackagePath(PathObjHolder{}),
		"function": "PingTest",
	}).Info("PingTest")

	return results
}

// PingTest gets roundTrip time to issue the "PING" command
func PingTest(s *Server) (results int64, error error) {
	conn, err := Dial(s.Host)
	if err != nil {
		return results, error
	}

	return Ping(conn, 5), err
}

// Version retrieves and parses the protocol version
func Version(conn net.Conn) (version string) {
	resp, err := command(conn, "HI")
	if err != nil {
		log.WithFields(log.Fields{
			"err":      err,
			"package":  GetPackagePath(PathObjHolder{}),
			"function": "Version",
		}).Debug()
	}
	verLine := strings.Split(resp, " ")
	version = verLine[1]
	return version
}

func calcMbps(start time.Time, finish time.Time, numBytes int64) (mbps float64) {
	diff := finish.Sub(start)
	secs := float64(diff.Nanoseconds()) / float64(1000000000)
	megabits := float64(numBytes) / float64(125000)
	mbps = megabits / secs

	return mbps
}
