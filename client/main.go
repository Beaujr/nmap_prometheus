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
)

func main() {
	log.Println("Application Starting")
	flag.Parse()
	c := reporter.NewReporter(*address)
	for !*ble {
		addresses, err := network.Scan(*subnet)
		if err != nil {
			log.Fatalf("unable to run nmap scan: %v", err)
		}
		err = c.Address(addresses)
		if err != nil {
			log.Panic(err)
		}
	}

	if *ble {
		err := bluetooth.Scan(&c)
		if err != nil {
			log.Printf("unable to run ble scan: %v", err)
		}
	}

}
