package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"strings"
	"time"
    "sync"
    "flag"
	"io"
	"net/http"
)
func main() {
	numConns := flag.Int("connections", 1, "Number of TCP connections or HTTP req in case of HTTP/2.0")
	dest := flag.String("dest", "127.0.0.1:8443", "Destination IP:port")
	path := flag.String("path", "/hostname", "Accessing path")
	protocol := flag.Float64("HTTPver", 1.1, "HTTP ver.")
    
	flag.Parse()

	switch *protocol {
	case 1.1:
	    fmt.Printf("Starting %d TCP connections to %s\n", *numConns, *dest)
		http11(*numConns, *dest, *path)
	case 2.0:
		fmt.Printf("Sending %d HTTP GET requests to %s\n", *numConns, *dest)
	    http20(*numConns, *dest, *path)
	default:
	    fmt.Println("HTTP ver. not supported. Select 1.1 or 2.0")
	}
}

func http11(ConnNums int, dest string, path string) {
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	
	var (
    HostCount = make(map[string]int)
    mu        sync.Mutex 
	wg sync.WaitGroup
	)

	for i := 0; i < ConnNums; i++ {
		wg.Add(1)
		go func (id int, dest string, path string) {
			defer wg.Done()
			conn, err := tls.Dial("tcp", dest, tlsConfig)
			if err != nil {
				fmt.Printf("[Conn %d] Failed: %v\n", id, err)
				return
			}

			// Activate this following liine if necessary while troubleshooting
			//fmt.Printf("[Conn: %d] established.\n", id)
			defer conn.Close()

			host := strings.Split(dest, ":")[0]
			req := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n",path, host)
			_ , err = conn.Write([]byte(req))

			if err != nil {
			  fmt.Printf("[Conn %d] Write error: %v\n", id, err)
			  return	
			}

			reader := bufio.NewReader(conn)
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			
		    if path != "/hostname" {
		        body, err := io.ReadAll(reader)
		        if err != nil {
		            fmt.Println("read error:", err)
		            return
		        }
		        fmt.Println(string(body))
		    
		    } else {
				for {
		    	  line, err := reader.ReadString('\n')
		    	  if err != nil {
		    	      if err == io.EOF { break }
		    	      fmt.Println("read error:", err)
		    	      break
		    	  }

		    	  cleanLine := strings.TrimSpace(line)

		    	  if strings.HasPrefix(cleanLine, "Hostname") {
		    	      mu.Lock()
		    	      HostCount[cleanLine]++
		    	      mu.Unlock()
		    	      break
		    	  }
		        }
			  } 
		}(i, dest, path)

	}
	wg.Wait()
	if path == "/hostname" {
	  for h, count := range HostCount {
			fmt.Printf("%s: %d\n", h, count)
		}
	}
}

func http20(ConnNums int, dest string, path string) {
	tr := &http.Transport{ TLSClientConfig: &tls.Config{ InsecureSkipVerify: true } }

	client := &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
	}

	var (
		hostCount = make(map[string]int)
		mu        sync.Mutex
		wg        sync.WaitGroup
	)
	url := "https://" + dest + path

	for i := 0; i < ConnNums; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			//client.GET() is simpler. But Client.Do() is easier to extend its functionality.
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				fmt.Printf("[Req %d] build error: %v\n", id, err)
				return
			}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("[Req %d] request error: %v\n", id, err)
				return
			}
			defer resp.Body.Close()

			if path != "/hostname" {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					fmt.Println("read error:", err)
					return
				}
				fmt.Println(string(body))
				return
			}

			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())

				if strings.HasPrefix(line, "Hostname") {
					mu.Lock()
					hostCount[line]++
					mu.Unlock()
					break
				}
			}

			if err := scanner.Err(); err != nil {
				fmt.Println("scan error:", err)
			}

		}(i)
	}

	wg.Wait()

	if path == "/hostname" {
		for h, count := range hostCount {
			fmt.Printf("%s: %d\n", h, count)
		}
	}
}