package agent

import (
	"context"
	"flag"
	"fmt"
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
	Scan() error
	ChkErr(error)
	AdvHandler(a ble.Advertisement)
}

type beaconScanner struct {
	BleScanner
	device ble.Device
}

// NewBeaconScanner returnings a new instance of a BleScanner
func NewBeaconScanner() (BleScanner, error) {
	d, err := dev.NewDevice(*device)
	if err != nil {
		return nil, err
	}
	ble.SetDefaultDevice(d)
	bls := beaconScanner{device: d}
	return &bls, nil
}

// Scan inits the HCI bluetooth and reports to the GRPC Server
func (bs *beaconScanner) Scan() error {
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), *du))
	bs.ChkErr(ble.Scan(ctx, *dup, bs.AdvHandler, nil))
	return nil
}

// AdvHandler is for handling Bluetooth Mac addresses
func (bs *beaconScanner) AdvHandler(a ble.Advertisement) {
	mac := a.Addr().String()
	log.Printf("Mac Address detected: %s, signal strength: %d", mac, a.TxPowerLevel())
}

// ChkErr controls the error channel while scanning ble
func (bs *beaconScanner) ChkErr(err error) {
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
