package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var (
	failed uint64
	good   uint64
)

func main() {
	concurrentRequest := 40
	updateGitCache := func(url string, w chan struct{}, wg *sync.WaitGroup) {
		defer wg.Done()

		jsonStr := []byte(fmt.Sprintf(`{"url":"http://%s"}`, url))

		client := NewTimeoutClient()
		resp, err := client.Post(os.Args[1], "application/json", bytes.NewBuffer(jsonStr))
		if err != nil {
			atomic.AddUint64(&failed, 1)
			println("Failed", url)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			atomic.AddUint64(&good, 1)
			println("Good", url)
		} else {
			atomic.AddUint64(&failed, 1)
			println("Failed", url)
		}
		<-w
	}
	wg := new(sync.WaitGroup)
	ws := make(chan struct{}, concurrentRequest)
	inFile, err := os.Open("./projects.txt")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer inFile.Close()

	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		ws <- struct{}{}
		wg.Add(1)
		go updateGitCache(scanner.Text(), ws, wg)
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
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(config.ReadWriteTimeout))
		return conn, nil
	}
}

func NewTimeoutClient() *http.Client {
	// Default configuration
	config := &Config{
		ConnectTimeout:   1 * time.Second,
		ReadWriteTimeout: 180 * time.Second,
	}

	return &http.Client{
		Transport: &http.Transport{
			Dial: TimeoutDialer(config),
		},
	}
}
