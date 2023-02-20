package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

func AWSRegionSubnet(region string) []string {
	// Get the AWS IP address ranges.
	resp, err := http.Get("https://ip-ranges.amazonaws.com/ip-ranges.json")
	if err != nil {
		log.Fatalf("Error getting AWS IP address ranges: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading AWS IP address ranges: %v", err)
	}

	// Parse the JSON object to get the IP address ranges for the specified region.
	type IPRange struct {
		IPPrefix string `json:"ip_prefix"`
		Region   string `json:"region"`
	}

	type AWSIPRanges struct {
		SyncToken    string    `json:"syncToken"`
		CreateDate   string    `json:"createDate"`
		Prefixes     []IPRange `json:"prefixes"`
		Ipv6Prefixes []IPRange `json:"ipv6_prefixes"`
	}

	var awsIPRanges AWSIPRanges
	err = json.Unmarshal(body, &awsIPRanges)
	if err != nil {
		log.Fatalf("Error parsing AWS IP address ranges: %v", err)
	}

	// Filter the IP address ranges by region.
	var filteredIPRanges []IPRange
	for _, ipRange := range awsIPRanges.Prefixes {
		if ipRange.Region == region {
			filteredIPRanges = append(filteredIPRanges, ipRange)
		}
	}

	// Generate a list of subnets to scan from the filtered IP address ranges.
	var subnets []string
	for _, ipRange := range filteredIPRanges {
		_, ipnet, err := net.ParseCIDR(ipRange.IPPrefix)
		if err != nil {
			log.Fatalf("Error parsing IP prefix: %v", err)
		}

		// Get the list of IP addresses in the subnet and add them to the list of
		// subnets to scan.
		ips, err := getHosts(ipnet.String())
		if err != nil {
			log.Fatalf("Error getting IP addresses for subnet %v: %v", ipnet, err)
		}

		subnets = append(subnets, ips...)
	}

	return subnets
}
