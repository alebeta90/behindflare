# Behindflare

This tool was created as a **Proof of Concept**  to reveal the threats related to web service misconfiguration using CloudFlare as reverse proxy and WAF.


## Problem

Most of CloudFlare users believe, that just setting up the reverse proxy which ensures security protection, will secure their back-end servers. This group of users are not aware that the attacker can find access to the back-end servers if he finds their IP addresses. There are plenty of passive and active techniques that can lead you to get the IP address of the Web App server.

## Service

If you would like to protect your servers against this kind of attack you can contact us at [Gonkar IT Security LTD](https://gonkar.com/gonkar-team-support/)  

# Usage


``` bash
./behindflare -h
Usage of ./behindflare:
  -domain string
    	Domain target (default "example.com")
  -jobs int
    	Number of parallel jobs (default 20)
  -proto string
    	The protocol used by the site behind CF (default "http")
  -subnet string
    	Subnet to scan (default "192.168.0.1/24")
``` 

# Disclaimer

This tool had been developed for research and educational purpose. Its usage for illegal actions is against creator will.
