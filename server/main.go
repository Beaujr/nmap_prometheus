package main

import (
	"flag"
	"fmt"
	"github.com/beaujr/nmap_prometheus/house"
	pb "github.com/beaujr/nmap_prometheus/proto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
)

var port = flag.String("port", "50051", "Port for GRPC Server")
var apiPort = flag.String("apiPort", "2112", "Port for API Server")

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	server := house.NewServer()
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
