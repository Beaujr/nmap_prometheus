package reporter

import (
	"context"
	"fmt"
	pb "github.com/beaujr/nmap_prometheus/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

// Reporter is the struct to handle GRP Comms
type Reporter struct {
	address string
	home    string
}

// NewReporter returns a Reporter for gRPC
func NewReporter(address string, home string) Reporter {
	return Reporter{address: address, home: home}
}

// Address reports pb.AddressRequest to the GRPC server
func (r *Reporter) Address(items []pb.AddressRequest) error {
	for _, item := range items {
		conn, err := grpc.Dial(r.address, grpc.WithInsecure())
		if err != nil {
			return err
		}
		defer func() {
			err = conn.Close()
		}()
		if err != nil {
			return err
		}
		c := pb.NewHomeDetectorClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*1000)
		defer cancel()
		item.Home = r.home
		response, err := c.Address(ctx, &item)
		if err != nil {
			grpcError := status.FromContextError(err)
			grpcErrorCode := grpcError.Code()
			if grpcErrorCode == codes.Unknown {
				fmt.Println("unable to talk to grpc server")
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
		conn, err := grpc.Dial(r.address, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer func() {
			err = conn.Close()
		}()
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		c := pb.NewHomeDetectorClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*1000)
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
