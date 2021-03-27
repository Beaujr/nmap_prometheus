package reporter

import (
	"context"
	"flag"
	pb "github.com/beaujr/nmap_prometheus/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"time"
)

var timeout = flag.Int("timeout", 10, "When to timeout connecting to server")
var netInterface = flag.String("interface", "", "Interface to bind to")

// Reporter is the struct to handle GRP Comms
type Reporter struct {
	home string
	conn *grpc.ClientConn
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
	return Reporter{home: home, conn: conn}
}

// Addresses reports pb.AddressesRequest to the GRPC server
func (r *Reporter) Addresses(items []*pb.AddressRequest) error {
	gAddr := pb.AddressesRequest{Addresses: items}
	c := pb.NewHomeDetectorClient(r.conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*(time.Duration(*timeout)))
	defer cancel()
	response, err := c.Addresses(ctx, &gAddr)
	if err != nil {
		grpcError := status.FromContextError(err)
		grpcErrorCode := grpcError.Code()
		if grpcErrorCode == codes.Unknown {
			log.Println("unable to talk to grpc server")
		}
		return err
	}
	log.Println(response.Acknowledged)
	return nil
}

// Address reports pb.AddressRequest to the GRPC server
func (r *Reporter) Address(items []*pb.AddressRequest) error {
	for _, item := range items {
		c := pb.NewHomeDetectorClient(r.conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*(time.Duration(*timeout)))
		defer cancel()
		item.Home = r.home
		response, err := c.Address(ctx, item)
		if err != nil {
			grpcError := status.FromContextError(err)
			grpcErrorCode := grpcError.Code()
			if grpcErrorCode == codes.Unknown {
				log.Println("unable to talk to grpc server")
			}
			return err
		}
		log.Println(response.Acknowledged)
	}
	return nil
}

// Bles is for handling Bluetooth Mac addresses
func (r *Reporter) Bles(macs []*string) error {
	for _, mac := range macs {
		c := pb.NewHomeDetectorClient(r.conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*(time.Duration(*timeout)))
		defer cancel()
		response, err := c.Ack(ctx, &pb.BleRequest{Mac: *mac, Home: r.home})
		if err != nil {
			return err
		}
		if response.Acknowledged {
			log.Printf("%s, %v", *mac, response.Acknowledged)
		}
	}
	return nil
}
