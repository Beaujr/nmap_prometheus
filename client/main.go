package main

import (
	"flag"
	"github.com/beaujr/nmap_prometheus/network"
	"github.com/beaujr/nmap_prometheus/reporter"
	"log"
)

var (
	subnet  = flag.String("subnet", "192.168.1.100-254", "NMAP subnet")
	address = flag.String("server", "192.168.1.190:50051", "NMAP subnet")
)

func main() {
	log.Println("Application Starting")
	flag.Parse()
	for true {
		addresses, err := network.Scan(*subnet)
		if err != nil {
			log.Fatalf("unable to run nmap scan: %v", err)
		}
		c := reporter.NewReporter(*address)
		err = c.Address(addresses)
		if err != nil {
			log.Panic(err)
		}
	}
}
