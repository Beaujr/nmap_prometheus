package network

import (
	"context"
	"github.com/Ullaakut/nmap"
	pb "github.com/beaujr/nmap_prometheus/proto"
	"log"
	"time"
)

// Scan executes the nmap binary and parses the result
func Scan(subnet string) ([]pb.AddressRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Equivalent to `/usr/local/bin/nmap -p 80,443,843 google.com facebook.com youtube.com`,
	// with a 5 minute timeout.
	scanner, err := nmap.NewScanner(
		nmap.WithTargets(subnet),
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

	// Use the results to print an example output
	addresses := make([]pb.AddressRequest, 0)
	for _, host := range result.Hosts {
		item := pb.AddressRequest{}
		for _, address := range host.Addresses {
			if address.AddrType == "ipv4" {
				item.Ip = address.Addr
			} else {
				item.Mac = address.Addr
			}
		}
		addresses = append(addresses, item)
	}
	return addresses, nil
}
