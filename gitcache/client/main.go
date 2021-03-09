package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"time"
)

var (
	failed uint64
	good   uint64
)

func main() {
	concurrentRequest := 10
	// tempDir is the downloading the projects.txt
	tempDir := os.Getenv("TEMP_DIR")
	if tempDir == "" {
		log.Panic("TEMP_DIR env is not set.")
	}

	url := os.Getenv("GITCACHE_URL")
	if url == "" {
		log.Panic("GITCACHE_URL env is not set.")
	}
	projecttxt := path.Join(tempDir, "projects.txt")
	txt, err := os.Create(projecttxt)
	if err != nil {
		log.Fatalf("unable to create projects.txt file %s", err.Error())
	}
	defer txt.Close()

	//nolint
	resp, err := http.Get("https://raw.githubusercontent.com/ossf/scorecard/main/cron/projects.txt")
	if err != nil {
		//nolint
		log.Fatalf("unable to download projects.txt file %s", err.Error())
	}
	defer resp.Body.Close()

	b, err := io.Copy(txt, resp.Body)
	if err != nil {
		log.Fatalf("unable to copy projects.txt file %s", err.Error())
	}
	fmt.Println("File size: ", b)

	updateGitCache := func(u string, w chan struct{}, wg *sync.WaitGroup) {
		defer wg.Done()

		jsonStr := []byte(fmt.Sprintf(`{"url":"http://%s"}`, u))

		client := NewTimeoutClient()
		//nolint
		resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonStr))
		if err != nil {
			atomic.AddUint64(&failed, 1)
			println("Failed", u, err.Error())
			<-w
			return
		}
		defer resp.Body.Close()

		const ok int = 200
		if resp.StatusCode == ok {
			atomic.AddUint64(&good, 1)
			println("Good", u)
		} else {
			atomic.AddUint64(&failed, 1)
			println("Failed", u, resp.Status)
		}
		<-w
	}
	wg := new(sync.WaitGroup)
	ws := make(chan struct{}, concurrentRequest)
	inFile, err := os.Open(projecttxt)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer inFile.Close()

	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		u := scanner.Text()
		ws <- struct{}{}
		wg.Add(1)
		go updateGitCache(u, ws, wg)
	}
	wg.Wait()
	fmt.Println("Total Number of failed request ", failed)
	fmt.Println("Total Number of good request ", good)
}

type Config struct {
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
}

func TimeoutDialer(config *Config) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, config.ConnectTimeout)
		if err != nil {
			//nolint
			return nil, err
		}
		//nolint
		conn.SetDeadline(time.Now().Add(config.ReadWriteTimeout))
		return conn, nil
	}
}

func NewTimeoutClient() *http.Client {
	// Default configuration
	config := &Config{
		ConnectTimeout: 1 * time.Second,
		//nolint
		ReadWriteTimeout: 180 * time.Second,
	}

	return &http.Client{
		Transport: &http.Transport{
			Dial: TimeoutDialer(config),
		},
	}
}
