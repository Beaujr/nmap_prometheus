package main

import (
	"flag"
	"github.com/beaujr/nmap_prometheus/agent"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"strings"
	"time"
)

var (
	subnet  = flag.String("subnet", "192.168.1.100-254", "NMAP subnet")
	address = flag.String("server", "192.168.1.190:50051", "NMAP Server")
	ble     = flag.Bool("ble", false, "Boolean for BLE scanning")
	home    = flag.String("home", "default", "Agent Location eg: Home, Dads house")
	bulk    = flag.Int("bulk", 10, "When to upload in bulk vs singular")
	script  = flag.Bool("script", false, "Set to true to run once off scan and report")
)

func main() {
	log.Println("Application Starting")
	flag.Parse()
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

	c := agent.NewReporter(*address, *home)
	if *ble {
		processBLE(&c)
	} else {
		processNMAP(&c, localAddresses)
	}

}
func processBLE(c *agent.Reporter) {
	err := c.Scan()
	if err != nil {
		grpcError := status.FromContextError(err)
		grpcErrorCode := grpcError.Code()
		if grpcErrorCode == codes.Unknown {
			log.Println("unable to talk to grpc server")
		}
		log.Printf("unable to run ble scan: %v", err)
		time.Sleep(2 * time.Second)
	}
}

func processNMAP(c *agent.Reporter, localAddresses map[string]string) {
	errors := 0
	for {
		addresses, err := c.ScanNmap(*subnet, *home, localAddresses)
		if err != nil {
			log.Printf("unable to run nmap scan: %v", err)
			errors++
		}

		if len(addresses) > *bulk {
			log.Printf("Bulk GRPC report: %d", len(addresses))
			err = c.Addresses(addresses)
			if err != nil {
				log.Printf("unable to run GRPC report: %v", err)
				time.Sleep(2 * time.Second)
				errors++
			} else {
				errors = 0
			}
		} else {
			err = c.Address(addresses)
			if err != nil {
				grpcError := status.FromContextError(err)
				grpcErrorCode := grpcError.Code()
				if grpcErrorCode == codes.Unknown {
					log.Println("unable to talk to grpc server")
				}
				log.Printf("unable to run GRPC report: %v", err)
				time.Sleep(2 * time.Second)
				errors++
			} else {
				errors = 0
			}
		}

		if *script {
			return
		}
		if errors >= 100 {
			log.Fatalf("Failed for last %d seconds", errors/2)
		}
	}
}
