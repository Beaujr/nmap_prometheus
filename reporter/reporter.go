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

type Reporter struct {
	address string
}

func NewReporter(address string) Reporter {
	return Reporter{address: address}
}

func (r *Reporter) Address(items []pb.AddressRequest) error {
	for _, item := range items {
		conn, err := grpc.Dial(r.address, grpc.WithInsecure())
		if err != nil {
			return err
		}
		defer conn.Close()
		c := pb.NewHomeDetectorClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*1000)
		defer cancel()
		log.Println(item)
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

func (r *Reporter) Bles(macs []*string) error {
	for _, mac := range macs {
		fmt.Println(*mac)
		conn, err := grpc.Dial(r.address, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c := pb.NewHomeDetectorClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*1000)
		defer cancel()
		response, err := c.Ack(ctx, &pb.BleRequest{Mac: *mac})
		if err != nil {
			return err
		}
		log.Println(response.Acknowledged)
	}
	return nil
}
