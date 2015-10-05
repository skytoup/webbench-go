package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
	"webbench-go/socket"
)

const (
	versionCode = "1.5"
)

var (
	force   = flag.Bool("f", false, "Don't wait for reply from server.")
	reload  = flag.Bool("r", false, "Send reload request - Pragma: no-cache.")
	sec     = flag.Int64("t", 30, "Run benchmark for <sec> seconds. Default 30.")
	port    = flag.Int("p", 80, "Use proxy server for request.")
	clients = flag.Int("c", 1, "Run <n> HTTP clients at once. Default one.")
	http09  = flag.Bool("9", false, "Use HTTP/0.9 style requests.")
	http10  = flag.Bool("1", false, "Use HTTP/1.0 protocol.")
	http11  = flag.Bool("2", false, "Use HTTP/1.1 protocol.")
	get     = flag.Bool("get", false, "Use GET request method.")
	head    = flag.Bool("head", false, "Use HEAD request method.")
	options = flag.Bool("options", false, "Use OPTIONS request method.")
	trace   = flag.Bool("trace", false, "Use TRACE request method.")
	help    = flag.Bool("h", false, "This information.")
	version = flag.Bool("V", false, "Display program version.")
)

var (
	request, method string
	httpVersion     = 1
)

func main() {
	flag.Parse()

	if *help {
		flag.Usage()
		return
	} else if *version {
		fmt.Println("webbench version" + versionCode)
		return
	} else if *get {
		method = "GET"
	} else if *options {
		method = "OPTIONS"
	} else if *head {
		method = "HEAD"
	} else if *trace {
		method = "TRACE"
	} else if len(method) == 0 {
		method = "GET"
	}

	host := flag.Arg(0)
	if len(host) == 0 {
		fmt.Println("webbench: Miss Url At Last!")
		return
	} else if strings.Index(host, "http://") == 0 {
		host = host[7:]
	}

	build_request(&host)

	fmt.Println("Webbench - Simple Web Benchmark" + versionCode)
	fmt.Println("Copyright (c) Radim Kolar 1997-2004, GPL Open Source Software.\n")
	fmt.Print("Benchmarking: ")

	fmt.Print(method, " ", host)

	if *http09 {
		fmt.Print(" (using HTTP/0.9)")
	} else if *http11 {
		fmt.Print(" (using HTTP/1.1)")
	}
	fmt.Print("\n")

	fmt.Printf("%d clients", *clients)

	fmt.Printf(", running %d sec", *sec)

	if *force {
		fmt.Print(", early socket close")
	}

	if *reload {
		fmt.Print(", forcing reload")
	}

	fmt.Print(".\n")
	// fmt.Println(*force, *reload, *sec, *port, *clients, *http09, *http10, *http11, *get, *head, *options, *trace, *help, *version)

	bench(&host)
}

func build_request(url *string) {
	req := bytes.NewBufferString(request)
	defer func() { request = req.String() }()

	if *http09 {
		httpVersion = 0
	} else if *http10 {
		httpVersion = 1
	} else if *http11 {
		httpVersion = 2
	}

	if *reload && httpVersion < 1 {
		httpVersion = 1
	} else if *head && httpVersion < 1 {
		httpVersion = 1
	} else if *options && httpVersion < 2 {
		httpVersion = 2
	} else if *trace && httpVersion < 2 {
		httpVersion = 2
	}

	req.WriteString(method)
	req.WriteByte(' ')

	if len(*url) > 1500 {
		fmt.Println("Url is too long.")
		os.Exit(2)
	}

	uri := "/"
	uriIdx := strings.Index(*url, "/")
	if uriIdx != -1 {
		uri = (*url)[uriIdx:]
	}
	req.WriteString(uri)

	switch httpVersion {
	case 1:
		req.WriteString(" HTTP/1.0")
	case 2:
		req.WriteString(" HTTP/1.1")
	}
	req.WriteString("\r\n")
	if httpVersion > 0 {
		req.WriteString("User-Agent: WebBench-go" + versionCode)
		req.WriteString("\r\n")
	}

	if httpVersion > 0 {
		req.WriteString("Host: ")
		if uriIdx != -1 {
			req.WriteString((*url)[:uriIdx])
		} else {
			req.WriteString(*url)
		}
		req.WriteString("\r\n")
	}

	if *reload {
		req.WriteString("Pragma: no-cache\r\n")
	}

	if httpVersion > 1 {
		req.WriteString("Connection: close\r\n")
	}

	if httpVersion > 0 {
		req.WriteString("\r\n")
	}

}

func bench(host *string) {
	var failS, bsS, speedS int64
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	cb := func(fail, bs, speed int64) {
		mutex.Lock()
		defer mutex.Unlock()
		failS += fail
		bsS += bs
		speedS += speed
		wg.Done()
	}
	for i := 0; i < *clients; i++ {
		wg.Add(1)
		bench_core(host, &request, *port, cb)
	}
	wg.Wait()

	ft := float64(*sec)
	fmt.Printf("\nSpeed=%d pages/min, %d bytes/sec.\nRequests: %d susceed, %d failed.\n", int64(float64(speedS+failS)/ft*60.0), int(float64(bsS)/ft), speedS, failS)
}

type bench_callback func(fail, bs, speed int64)

func bench_core(host, req *string, port int, callback bench_callback) {
	go benchcore(host, req, port, callback)
}

func benchcore(host, req *string, port int, callback bench_callback) {
	var fail, bs, speed int64

	buf := make([]byte, 1500)
	reqLen := len(*req)
	var conn *net.TCPConn

	ds := fmt.Sprintf("%ds", *sec)

	dur, _ := time.ParseDuration(ds)
	timer := time.NewTimer(dur)

	for {
	nextconn:
		select {
		case <-timer.C:
			callback(fail, bs, speed)
			fmt.Println("Some of our childrens died.")
			return
		default:
			conn = socket.Socket(host, port)
			if conn == nil {
				fail++
				continue
			}
			wlen, err := conn.Write([]byte(*req))
			if wlen != reqLen || err != nil {
				fail++
				conn.Close()
				continue
			}
			if httpVersion == 0 {
				err := conn.CloseWrite()
				if err != nil {
					fail++
					conn.Close()
					continue
				}
			}
			if *force == false {
				for {
					b, err := conn.Read(buf)
					if err == io.EOF {
						break
					} else if err != nil {
						fail++
						conn.Close()
						break nextconn
					}
					bs += int64(b)
				}
			}
			if conn.Close() != nil {
				fail++
				continue
			}
			speed++
		}
	}
}
