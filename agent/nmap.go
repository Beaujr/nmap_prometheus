package agent

import (
	"context"
	"github.com/Ullaakut/nmap"
	pb "github.com/beaujr/nmap_prometheus/proto"
	"log"
	"net"
	"strings"
	"time"
)

// GoogleAssistant interface for calling smart home api
type NetScanner interface {
	Scan() ([]*pb.AddressRequest, error)
}

// NewAssistant returns a new assistant client
func NewScanner(home string, subnet string) NetScanner {
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
					localAddresses[ip.String()] = hwAddress
				}
			}
		}
	}
	return &NetworkScanner{home: home, subnet: subnet, localAddrs: localAddresses}
}

// AssistantRelay is an implementation of the GoogleAssistant
type NetworkScanner struct {
	NetScanner
	home       string
	subnet     string
	localAddrs map[string]string
}

// Scan executes the nmap binary and parses the result
func (ns *NetworkScanner) Scan() ([]*pb.AddressRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Equivalent to `/usr/local/bin/nmap -p 80,443,843 google.com facebook.com youtube.com`,
	// with a 5 minute timeout.
	scanner, err := nmap.NewScanner(
		nmap.WithTargets(ns.subnet),
		nmap.WithPingScan(),
		nmap.WithContext(ctx),
	)
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
			if address.AddrType == "ipv4" {
				item.Ip = address.Addr
				if val, ok := ns.localAddrs[address.Addr]; ok {
					item.Mac = val
				}
				continue
			} else {
				item.Mac = address.Addr
			}
		}
		item.Home = ns.home
		addresses = append(addresses, &item)
	}
	return addresses, nil
}
