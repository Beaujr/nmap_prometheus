package agent

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/beaujr/nmap_prometheus/proto"
	"github.com/go-ble/ble"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"math"
	"net"
	"time"
)

var (
	bulk         = flag.Int("bulk", 10, "When to upload in bulk vs singular")
	script       = flag.Bool("script", false, "Set to true to run once off scan and report")
	subnet       = flag.String("subnet", "192.168.1.100-254", "NMAP subnet")
	address      = flag.String("server", "192.168.1.190:50051", "NMAP Server")
	bleEnabled   = flag.Bool("ble", false, "Boolean for BLE scanning")
	home         = flag.String("home", "default", "Agent Location eg: Home, Dads house")
	timeout      = flag.Int("timeout", 10, "When to timeout connecting to server")
	netInterface = flag.String("interface", "", "Interface to bind to")
	agentId      = flag.String("agentId", "nmapAgent", "Identify Agent, if left blank will be the Machines ID")
	apiKey       = flag.String("apikey", "apikey", "API KEY for access")
)

// Reporter is the struct to handle GRP Comms
type Reporter struct {
	BleScanner
	Home       string
	id         string
	conn       *grpc.ClientConn
	ignoreList map[string]bool
	Nmap       NetScanner
}

func dial(address string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if *netInterface != "" {
		localAddrDialier := &net.Dialer{
			LocalAddr: &net.TCPAddr{
				IP:   net.ParseIP(*netInterface),
				Port: 0,
			},
		}
		conn, err = grpc.Dial(address, grpc.WithInsecure(), grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return localAddrDialier.DialContext(ctx, "tcp", addr)
		}))

	}
	return conn, err
}

// NewReporter returns a Reporter for gRPC
func NewReporter() Reporter {
	conn, err := dial(*address)
	if err != nil {
		log.Print(err)
	}
	ignoreList := make(map[string]bool)
	if *bleEnabled {
		bls, err := NewBeaconScanner()
		if err != nil {
			log.Print(err)
		}
		return Reporter{BleScanner: bls, Home: *home, conn: conn, id: *agentId, ignoreList: ignoreList, Nmap: nil}
	}
	nmapScanner := NewScanner(*home, *subnet)
	return Reporter{BleScanner: nil, Home: *home, conn: conn, id: *agentId, ignoreList: ignoreList, Nmap: nmapScanner}

}
func (r *Reporter) buildClient() (pb.HomeDetectorClient, context.Context, context.CancelFunc) {
	client := pb.NewHomeDetectorClient(r.conn)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*(time.Duration(*timeout)))
	ctx = metadata.AppendToOutgoingContext(ctx, "client", r.id)
	ctx = metadata.AppendToOutgoingContext(ctx, "home", r.Home)
	ctx = metadata.AppendToOutgoingContext(ctx, "apikey", *apiKey)
	return client, ctx, cancelFunc
}

// Addresses reports pb.AddressesRequest to the GRPC server
func (r *Reporter) Addresses(items []*pb.AddressRequest) error {
	gAddr := pb.AddressesRequest{Addresses: items}
	c, ctx, cancel := r.buildClient()
	defer cancel()
	response, err := c.Addresses(ctx, &gAddr)
	if err != nil {
		return err
	}
	log.Println(response.Acknowledged)
	return nil
}

// Address reports pb.AddressRequest to the GRPC server
func (r *Reporter) Address(items []*pb.AddressRequest) error {
	c, ctx, cancel := r.buildClient()
	defer cancel()
	for _, item := range items {
		response, err := c.Address(ctx, item)
		if err != nil {
			return err
		}
		log.Println(response.Acknowledged)
	}
	return nil
}

// AdvHandler is for handling Bluetooth Mac addresses while scanning
func (r *Reporter) AdvHandler(a ble.Advertisement) {
	mac := a.Addr().String()
	numerator := -69 - a.RSSI()
	distance := float64(math.Pow(10, float64(numerator)/float64(10)))
	log.Printf("Mac: %s, Distance: %v RSSI: %d\n", mac, distance, a.RSSI())
	if val, ok := r.ignoreList[mac]; ok && !val {
		log.Println(fmt.Sprintf("Not reporting ble: %s", mac))
		return
	}
	c, ctx, cancel := r.buildClient()
	defer cancel()
	response, err := c.Ack(ctx, &pb.BleRequest{Key: mac, Distance: float32(distance)})
	if err != nil {
		log.Println(fmt.Sprintf("GRPC Error: %s", err.Error()))
		return
	}
	if response.Acknowledged {
		log.Printf("%s, %v", mac, response.Acknowledged)
	}
	r.ignoreList[mac] = response.Acknowledged
}

// ProcessNMAP scans the network and reports to nmap server
func (r *Reporter) ProcessNMAP() {
	errors := 0
	for {
		addresses, err := r.Nmap.Scan()
		if err != nil {
			log.Printf("unable to run nmap scan: %v", err)
			errors++
		}
		//addresses := make([]*pb.AddressRequest, 0)
		//addresses = append(addresses, &pb.AddressRequest{Mac: "0000", Ip: "192.168.16.2"})
		//err := fmt.Errorf("not a real error")
		if len(addresses) > *bulk {
			err := r.bulkReport(addresses)
			if err != nil {
				log.Printf("unable to run GRPC report: %v", err)
				time.Sleep(2 * time.Second)
				errors++
			} else {
				errors = 0
			}
		} else {
			err = r.Address(addresses)
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

func (r *Reporter) bulkReport(addresses []*pb.AddressRequest) error {
	log.Printf("Bulk GRPC report: %d", len(addresses))
	err := r.Addresses(addresses)
	if err != nil {
		log.Printf("unable to run GRPC report: %v", err)
		return err
	}
	return nil
}

func (r *Reporter) singleReport(addresses []*pb.AddressRequest) error {
	log.Printf("Bulk GRPC report: %d", len(addresses))
	err := r.Addresses(addresses)
	if err != nil {
		log.Printf("unable to run GRPC report: %v", err)
		return err
	}
	return nil
}

// Scan inits the HCI bluetooth and reports to the GRPC Server
func (r *Reporter) Scan() error {
	// Scan for specified durantion, or until interrupted by user.
	log.Printf("Scanning for %s...\n", *du)
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), *du))
	r.ChkErr(ble.Scan(ctx, *dup, r.AdvHandler, nil))
	return nil
}

// ProcessBLE scans the Bluetooth and reports to the grpc server
func (r *Reporter) ProcessBLE() {
	err := r.Scan()
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
