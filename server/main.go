package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/beaujr/nmap_prometheus/house"
	pb "github.com/beaujr/nmap_prometheus/proto"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
)

var port = flag.String("port", "50051", "Port for GRPC Server")
var apiPort = flag.String("apiPort", "2112", "Port for API Server")

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	server := house.NewServer(ctx)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	pb.RegisterHomeDetectorServer(s, server.(pb.HomeDetectorServer))
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/devices", server.Devices)
	http.HandleFunc("/people", server.People)
	http.HandleFunc("/empty", server.HomeEmptyState)
	go http.ListenAndServe(fmt.Sprintf(":%s", *apiPort), nil)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
