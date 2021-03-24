package main

import (
	"flag"
	"github.com/beaujr/nmap_prometheus/bluetooth"
	"github.com/beaujr/nmap_prometheus/network"
	"github.com/beaujr/nmap_prometheus/reporter"
	"log"
)

var (
	subnet  = flag.String("subnet", "192.168.1.100-254", "NMAP subnet")
	address = flag.String("server", "192.168.1.190:50051", "NMAP Server")
	ble     = flag.Bool("ble", false, "Boolean for BLE scanning")
	home    = flag.String("home", "default", "Agent Location eg: Home, Dads house")
)

func main() {
	log.Println("Application Starting")
	flag.Parse()
	c := reporter.NewReporter(*address, *home)
	for !*ble {
		addresses, err := network.Scan(*subnet)
		if err != nil {
			log.Printf("unable to run nmap scan: %v", err)
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
