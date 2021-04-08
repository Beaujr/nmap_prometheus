package main

import (
	"flag"
	"github.com/beaujr/nmap_prometheus/bluetooth"
	"github.com/beaujr/nmap_prometheus/network"
	"github.com/beaujr/nmap_prometheus/reporter"
	"log"
	"net"
)

var (
	subnet  = flag.String("subnet", "192.168.1.100-254", "NMAP subnet")
	address = flag.String("server", "192.168.1.190:50051", "NMAP Server")
	ble     = flag.Bool("ble", false, "Boolean for BLE scanning")
	home    = flag.String("home", "default", "Agent Location eg: Home, Dads house")
	bulk    = flag.Int("bulk", 10, "When to upload in bulk vs singular")
)

func main() {
	log.Println("Application Starting")
	flag.Parse()
	c := reporter.NewReporter(*address, *home)
	localAddresses := make(map[string]string)
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Println(err)
	}
	// handle err
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		// handle err

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			localAddresses[ip.String()] = i.HardwareAddr.String()
			// process IP address
		}
	}
	for !*ble {
		addresses, err := network.Scan(*subnet, *home, localAddresses)
		if err != nil {
			log.Printf("unable to run nmap scan: %v", err)
		}
		if len(addresses) > *bulk {
			log.Printf("Bulk GRPC report: %d", len(addresses))
			err = c.Addresses(addresses)
			if err != nil {
				log.Printf("unable to run GRPC report: %v", err)
			}
			continue
		}
		err = c.Address(addresses)
		if err != nil {
			log.Printf("unable to run GRPC report: %v", err)
		}
	}

	if *ble {
		err := bluetooth.Scan(&c)
		if err != nil {
			log.Printf("unable to run ble scan: %v", err)
		}
	}

}
