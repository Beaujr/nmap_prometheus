package main

import (
	"flag"
	ag "github.com/beaujr/nmap_prometheus/agent"
	"log"
)

func main() {
	log.Println("Application Starting")
	flag.Parse()
	c := ag.NewReporter()
	if c.Nmap != nil {
		c.ProcessNMAP()
	} else {
		c.ProcessBLE()
	}

}
