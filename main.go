package main

import (
	"crypto/tls"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
)

var (
	protocol      = flag.String("proto", "http", "The protocol used by the site behind CF")
	domain        = flag.String("domain", "example.com", "Domain target")
	subnet        = flag.String("subnet", "192.168.0.1/24", "Subnet to scan")
	OriginalTitle string
	limit         = flag.Int("jobs", 20, "Number of parallel jobs")
)

func main() {
	flag.Parse()

	Banner()

	color.Cyan("Analyzing Domain: %v", *domain)

	siteInfo()

	ipAddresses2, err := Hosts(*subnet)
	if err != nil {
		log.Fatal(err)
	}

	color.Cyan("Number of IPs to scan: %v", len(ipAddresses2))
	// The `main` func must not finished before all jobs are done. We use a
	// WaitGroup to wait for all of them.
	wg := new(sync.WaitGroup)

	// We use a buffered channel as a semaphore to limit the number of
	// concurrently executing jobs.
	sem := make(chan struct{}, *limit)

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

// Get Body Size
func siteInfo() {

	resp, err := http.Get("https://" + *domain)
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading body", err)
	}

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(body)))

	OriginalTitle = doc.Find("title").Eq(0).Text()

}

func scanBlock(k int, i string) string {

	req, err := http.NewRequest("GET", *protocol+"://"+i, nil)
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}

	//fmt.Println("Job Number: ", j, "Using IP: ", ips)
	//fmt.Println("Using IP: ", i)

	req.Host = *domain

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

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(body)))

	title := doc.Find("title").Eq(0).Text()

	switch OriginalTitle {
	case title:
		color.Green("##############-HOST FOUND-###################\n")
		color.Green("Server IP: %v", i)
		color.Green("HTTP Status: %v", resp.StatusCode)
		color.Green("#############################################\n")
		defer resp.Body.Close()

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
