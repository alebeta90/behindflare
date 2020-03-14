package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
)

var domain = os.Args[1]
var subnet = os.Args[2]
var bodySize int
var jobcount = 1000
var limit = 50

func main() {
	fmt.Println("Analyzing Domain: ", domain)

	siteInfo()

	ipAddresses2, err := Hosts(subnet)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Number of IPs to scan: ", len(ipAddresses2))
	// The `main` func must not finished before all jobs are done. We use a
	// WaitGroup to wait for all of them.
	wg := new(sync.WaitGroup)

	// We use a buffered channel as a semaphore to limit the number of
	// concurrently executing jobs.
	sem := make(chan struct{}, limit)

	// We run each job in its own goroutine but use the semaphore to limit
	// their concurrent execution.
	for k, i := range ipAddresses2 {
		// This job must be waited for.
		wg.Add(1)

		// Acquire the semaphore by writing to the buffered channel. If the
		// channel is full, this call will block until another job has released
		// it.
		sem <- struct{}{}

		// Now we have acquired the semaphore and can start a goroutine for
		// this job. Note that we must capture `i` as an argument.
		go func(k int, i string) {
			// When the work of this goroutine has been done, we decrement the
			// WaitGroup.
			defer wg.Done()

			// When the work of this goroutine has been done, we release the
			// semaphore.
			defer func() { <-sem }()
			// Do the actual work.
			scanBlock(k, i)
			//fmt.Printf("IP scanned is %s\n", result)

		}(k, i)
	}

	// Wait for all jobs to finish.
	wg.Wait()

}

func scanBlock(k int, i string) string {

	req, err := http.NewRequest("GET", "https://"+i, nil)
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}

	//fmt.Println("Job Number: ", j, "Using IP: ", ips)
	//fmt.Println("Using IP: ", i)

	req.Host = domain

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	resp, err := client.Do(req)
	if err != nil {
		//fmt.Println("HTTP call failed:", err)
		return i
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading body", err)
	}

	if resp.StatusCode == 200 && len(body) == bodySize {

		fmt.Printf("##############-HOST FOUND-###################\n")
		fmt.Println("Server IP: ", i)
		fmt.Println("HTTP Status: ", resp.StatusCode)
		fmt.Println("Body Length: ", len(body))
		fmt.Printf("#############################################\n")
		defer resp.Body.Close()
		//break
	}
	return i

}

// Hosts function read a CIDR and create a slide with all the IP addresses contained in it
func Hosts(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	// remove network address and broadcast address
	lenIPs := len(ips)
	switch {
	case lenIPs < 2:
		return ips, nil

	default:
		return ips[1 : len(ips)-1], nil
	}
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// Get Body Size
func siteInfo() {

	req, err := http.NewRequest("GET", "https://"+domain, nil)
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}

	client := http.DefaultClient

	resp, err := client.Do(req)
	if err != nil {
		panic(err)

	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading body", err)
	}

	fmt.Println("Site Body Length: ", len(body))
	bodySize = len(body)

	defer resp.Body.Close()

}
