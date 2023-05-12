package agent

import (
	"context"
	"github.com/Ullaakut/nmap"
	pb "github.com/beaujr/nmap_prometheus/proto"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

// NetScanner interface for Scanning and returning AddressRequests
type NetScanner interface {
	Scan() ([]*pb.AddressRequest, error)
}

// NewScanner returns a new NetScanner client
func NewScanner() NetScanner {
	localAddresses := make(map[string]string)
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Println(err)
	}
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		if err != nil {
			log.Println(err)
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip = addr.(*net.IPNet).IP.To4(); ip != nil {
				hwAddress := strings.ToUpper(i.HardwareAddr.String())
				if len(hwAddress) > 0 {
					log.Printf("Local Interface Mac (%s) Ip (%s)", hwAddress, ip.String())
					localAddresses[ip.String()] = hwAddress
				}
			}
		}
	}
	opts := []func(scanner *nmap.Scanner){
		nmap.WithTargets(*subnet),
		nmap.WithPingScan(),
	}
	if len(*dnsServers) > 0 {
		opts = append(opts, nmap.WithCustomDNSServers(strings.Split(*dnsServers, ",")...))
	} else {
		opts = append(opts, nmap.WithSystemDNS())
	}
	return &NetworkScanner{home: *Home, subnet: *subnet, localAddrs: localAddresses, options: opts}
}

// NetworkScanner is an implementation of the NetScanner
type NetworkScanner struct {
	NetScanner
	home       string
	subnet     string
	localAddrs map[string]string
	options    []func(scanner *nmap.Scanner)
}

// Scan executes the nmap binary and parses the result
func (ns *NetworkScanner) Scan() ([]*pb.AddressRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Equivalent to `/usr/local/bin/nmap -p 80,443,843 google.com facebook.com youtube.com`,
	// with a 5 minute timeout.
	opts := append(ns.options, nmap.WithContext(ctx))
	scanner, err := nmap.NewScanner(opts...)
	//scanner, err := nmap.NewScanner(
	//	nmap.WithTargets(ns.subnet),
	//	nmap.WithPingScan(),
	//	nmap.WithContext(ctx),
	//	nmap.WithSystemDNS(),
	//	//nmap.WithCustomDNSServers(),
	//)
	if err != nil {
		return nil, err
	}

	result, warnings, err := scanner.Run()
	if err != nil {
		return nil, err
	}
	if warnings != nil {
		log.Printf("Warnings: \n %v", warnings)
	}

	// Use the results to print an example output
	addresses := make([]*pb.AddressRequest, 0)
	for _, host := range result.Hosts {
		item := pb.AddressRequest{}
		for _, address := range host.Addresses {
			switch address.AddrType {
			case "ipv4":
				item.Ip = address.Addr
				if val, ok := ns.localAddrs[address.Addr]; ok {
					item.Mac = val
				}
				continue
			case "mac":
				item.Mac = address.Addr
				item.Vendor = address.Vendor
			default:
				item.Mac = address.Addr
			}
		}
		for _, hostnames := range host.Hostnames {
			item.Hosts = append(item.Hosts, hostnames.Name)
		}
		rstt, err := strconv.Atoi(host.Times.SRTT)
		if err != nil {
			rstt = 1000
		}
		item.Distance = float32(rstt / 1000)
		addresses = append(addresses, &item)
	}
	return addresses, nil
}
