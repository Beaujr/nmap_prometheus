package agent

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/beaujr/nmap_prometheus/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"net"
	"time"
)

var (
	timeout      = flag.Int("timeout", 10, "When to timeout connecting to server")
	netInterface = flag.String("interface", "", "Interface to bind to")
	agentId      = flag.String("agentId", "nmapAgent", "Identify Agent, if left blank will be the Machines ID")
)

// Reporter is the struct to handle GRP Comms
type Reporter struct {
	Home       string
	id         string
	conn       *grpc.ClientConn
	ignoreList map[string]bool
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
func NewReporter(address string, home string) Reporter {
	conn, err := dial(address)
	if err != nil {
		log.Print(err)
	}
	ignoreList := make(map[string]bool)
	return Reporter{Home: home, conn: conn, id: *agentId, ignoreList: ignoreList}
}
func (r *Reporter) buildClient() (pb.HomeDetectorClient, context.Context, context.CancelFunc) {
	client := pb.NewHomeDetectorClient(r.conn)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*(time.Duration(*timeout)))
	ctx = metadata.AppendToOutgoingContext(ctx, "client", r.id)
	ctx = metadata.AppendToOutgoingContext(ctx, "home", r.Home)
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
	for _, item := range items {
		c, ctx, cancel := r.buildClient()
		defer cancel()
		item.Home = r.Home
		response, err := c.Address(ctx, item)
		if err != nil {
			return err
		}
		log.Println(response.Acknowledged)
	}
	return nil
}

// Bles is for handling Bluetooth Mac addresses
func (r *Reporter) Bles(macs []*string) error {
	for _, mac := range macs {
		if val, ok := r.ignoreList[*mac]; ok && !val {
			log.Println(fmt.Sprintf("ignoring ble: %s", *mac))
			return nil
		}
		c, ctx, cancel := r.buildClient()
		defer cancel()
		response, err := c.Ack(ctx, &pb.BleRequest{Mac: *mac, Home: r.Home})
		if err != nil {
			return err
		}
		if response.Acknowledged {
			log.Printf("%s, %v", *mac, response.Acknowledged)
		}
		r.ignoreList[*mac] = response.Acknowledged
	}
	return nil
}
