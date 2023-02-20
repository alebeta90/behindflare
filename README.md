# Behindflare

This tool was created as a **Proof of Concept**  to reveal the threats related to web service misconfiguration using CloudFlare as reverse proxy and WAF.

## Problem

Most of CloudFlare users believe, that just setting up the reverse proxy which ensures security protection, will secure their back-end servers. This group of users are not aware that the attacker can find access to the back-end servers if he finds their IP addresses. There are plenty of passive and active techniques that can lead you to get the IP address of the Web App server.

## Service

If you would like to protect your servers against this kind of attack you can contact us at [Gonkar IT Security](https://gonkar.com/)  

## ToDo

* Azure Subnets
* GCE Subnets
* DO subnets

# Usage

Clone the repository and build the Golang binary with:  

`go build`

You should end up with a `behindflare` binary.  


``` bash
./behindflare -h
Usage of ./behindflare:
  -domain string
    	Domain target (default "example.com")
  -jobs int
    	Number of parallel jobs (default 20)
  -proto string
    	The protocol used by the site behind CF (default "http")
  -region string
  	AWS region to scan (optional)
  -subnet string
    	Subnet to scan (default "192.168.0.1/24")
``` 

To scan a subnet you can run the binary as follow:

`./behindflare -proto https -domain example.com -subnet 192.168.0.1/24 -jobs 50`

To scan all the subnets in the us-east-1 region for hosts that are behind Cloudflare, run the following command:  

`./behindflare -proto https -domain example.com -aws -region us-east-1 -jobs 50`

You can specify different AWS regions like `eu-central-1` for example.  

# Disclaimer

This tool had been developed for research and educational purpose. Its usage for illegal actions is against creator will.
