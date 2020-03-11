# Behindflare

This tool was created to demostrate the threats of misconfiguration of web services using CloudFlare as reverse proxy and WAF.  

# Usage

1. Change `domain` variable to the hostname you want to check
2. Change `ipAddresses2` variable to the subnet you decide to scan, e.g: Vultr, Godaddy, DigitalOcean, etc.

### Run it

`go run main.go`