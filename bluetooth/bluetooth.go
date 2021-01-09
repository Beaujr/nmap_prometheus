package bluetooth

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/beaujr/nmap_prometheus/proto"
	reporter "github.com/beaujr/nmap_prometheus/reporter"
	"log"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
	"github.com/pkg/errors"
)

var (
	device = flag.String("device", "default", "implementation of ble")
	du     = flag.Duration("du", 0*time.Second, "scanning duration")
	dup    = flag.Bool("dup", true, "allow duplicate reported")
)

// BleScanner Interface for BleScanner structs for BLE/BL
type BleScanner interface {
	report([]pb.AddressRequest, error)
}
type beaconScanner struct {
	*reporter.Reporter
}

// Scan inits the HCI bluetooth and reports to the GRPC Server
func Scan(reporter *reporter.Reporter) error {
	d, err := dev.NewDevice(*device)
	if err != nil {
		return err
	}
	ble.SetDefaultDevice(d)
	bs := beaconScanner{reporter}
	// Scan for specified durantion, or until interrupted by user.
	fmt.Printf("Scanning for %s...\n", *du)
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), *du))
	chkErr(ble.Scan(ctx, *dup, bs.advHandler, nil))
	return nil
}

func (b *beaconScanner) advHandler(a ble.Advertisement) {
	addresses := make([]*string, 0)
	mac := a.Addr().String()
	addresses = append(addresses, &mac)
	err := b.Reporter.Bles(addresses)
	if err != nil {
		log.Panic(err)
	}
	//if a.Connectable() {
	//	fmt.Printf("[%s] C %3d:", a.Addr(), a.RSSI())
	//} else {
	//	fmt.Printf("[%s] N %3d:", a.Addr(), a.RSSI())
	//}
	//comma := ""
	//if len(a.LocalName()) > 0 {
	//	fmt.Printf(" Name: %s", a.LocalName())
	//	comma = ","
	//}
	//if len(a.Services()) > 0 {
	//	fmt.Printf("%s Svcs: %v", comma, a.Services())
	//	comma = ","
	//}
	//if len(a.ManufacturerData()) > 0 {
	//	fmt.Printf("%s MD: %X", comma, a.ManufacturerData())
	//}
	//fmt.Printf("\n")
}

func chkErr(err error) {
	switch errors.Cause(err) {
	case nil:
	case context.DeadlineExceeded:
		fmt.Printf("done\n")
	case context.Canceled:
		fmt.Printf("canceled\n")
	default:
		log.Printf(err.Error())
	}
}
