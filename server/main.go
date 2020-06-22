package main

import (
	"flag"
	"github.com/beaujr/nmap_prometheus/house"
	pb "github.com/beaujr/nmap_prometheus/proto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
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
	pb.RegisterHomeDetectorServer(s, server.(pb.HomeDetectorServer))
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":2112", nil)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
