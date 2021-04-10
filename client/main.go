package main

import (
	"flag"
	"github.com/beaujr/nmap_prometheus/bluetooth"
	"github.com/beaujr/nmap_prometheus/network"
	"github.com/beaujr/nmap_prometheus/reporter"
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

	c := reporter.NewReporter(*address, *home)
	if *ble {
		processBLE(&c)
	} else {
		processNMAP(&c, localAddresses)
	}

}
func processBLE(c *reporter.Reporter) {
	err := bluetooth.Scan(c)
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

func processNMAP(c *reporter.Reporter, localAddresses map[string]string) {
	errors := 0
	for {
		addresses, err := network.Scan(*subnet, *home, localAddresses)
		//addresses := make([]*pb.AddressRequest, 0)
		//addresses = append(addresses, &pb.AddressRequest{Mac: "0000", Ip: "192.168.16.2"})
		//if err != nil {
		//	log.Printf("unable to run nmap scan: %v", err)
		//}
		if len(addresses) > *bulk {
			log.Printf("Bulk GRPC report: %d", len(addresses))
			err := c.Addresses(addresses)
			if err != nil {
				log.Printf("unable to run GRPC report: %v", err)
				time.Sleep(2 * time.Second)
				errors++
				continue
			}
			errors = 0
			continue
		}
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
		}
		if errors >= 100 {
			log.Fatalf("Failed for last %d seconds", errors/2)
		}
	}
}
