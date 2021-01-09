package main

import (
	"flag"
	"github.com/beaujr/nmap_prometheus/house"
	pb "github.com/beaujr/nmap_prometheus/proto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
)

const (
	port = ":50051"
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", port)
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
	go http.ListenAndServe(":2112", nil)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
