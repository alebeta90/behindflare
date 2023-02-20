package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
)

func main() {
	var protocol, domain, subnet string
	var limit int
	var mu sync.Mutex
	var numScanned int
	flag.StringVar(&protocol, "proto", "http", "The protocol used by the site behind CF")
	flag.StringVar(&domain, "domain", "example.com", "Domain target")
	flag.StringVar(&subnet, "subnet", "192.168.0.1/24", "Subnet to scan")
	flag.IntVar(&limit, "jobs", 20, "Number of parallel jobs")
	flag.Parse()

	Banner()

	color.Cyan("Analyzing Domain: %v", domain)

	originalTitle, err := getSiteTitle(protocol, domain)
	if err != nil {
		log.Fatalf("Error reading site title: %v", err)
	}

	ipAddresses, err := getHosts(subnet)
	if err != nil {
		log.Fatalf("Error getting IP addresses: %v", err)
	}

	color.Cyan("Number of IPs to scan: %v", len(ipAddresses))

	// The `main` func must not finish before all jobs are done. We use a
	// WaitGroup to wait for all of them.
	wg := new(sync.WaitGroup)

	// We use a buffered channel as a semaphore to limit the number of
	// concurrently executing jobs.
	sem := make(chan struct{}, limit)

	// We run each job in its own goroutine but use the semaphore to limit
	// their concurrent execution.
	for k, i := range ipAddresses {
		// This job must be waited for.
		wg.Add(1)

		// Acquire the semaphore by writing to the buffered channel. If the
		// channel is full, this call will block until another job has released
		// it.
		sem <- struct{}{}

		// Now we have acquired the semaphore and can start a goroutine for
		// this job. Note that we must capture `i` as an argument.
		go func(k int, i string) {
			defer func() { <-sem }()

			// When the work of this goroutine has been done, we decrement the
			// WaitGroup.
			defer wg.Done()

			// Do the actual work.
			if err := scanHost(k, i, protocol, domain, originalTitle); err != nil {
				log.Printf("Error scanning host %s: %v", i, err)
			}

			// Update the counter and print progress.
			mu.Lock()
			numScanned++
			if numScanned%100 == 0 {
				color.Cyan("Scanned %d hosts", numScanned)
			}
			mu.Unlock()
		}(k, i)
	}

	// Wait for all jobs to finish.
	wg.Wait()
}

func getSiteTitle(protocol, domain string) (string, error) {
	resp, err := http.Get(protocol + "://" + domain)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(body)))

	return doc.Find("title").Eq(0).Text(), nil
}

func getHosts(cidr string) ([]string, error) {
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

func scanHost(k int, ip, protocol, domain, originalTitle string) error {
	req, err := http.NewRequest("GET", protocol+"://"+ip, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Host = domain

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP call failed: %v", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading body: %v", err)
	}

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	title := doc.Find("title").Eq(0).Text()

	switch originalTitle {
	case title:
		color.Green("##############-HOST FOUND-###################\n")
		color.Green("Server IP: %v", ip)
		color.Green("HTTP Status: %v", resp.StatusCode)
		color.Green("#############################################\n")
	}

	return nil
}
