package house

import (
	"context"
	"fmt"
	pb "github.com/beaujr/nmap_prometheus/proto"
	"github.com/golang/protobuf/ptypes/empty"
	etcdv3 "github.com/ozonru/etcd/v3/clientv3"
	"log"
	"sort"
	"time"
)

// Ack for bluetooth reported MAC addresses
func (s *Server) Ack(ctx context.Context, in *pb.BleRequest) (*pb.Reply, error) {
	s.grpcPrometheusMetrics(ctx, "grpc_ble", "Ack")
	s.grpcHitsMetrics("grpc_address_count_ble", "Ack", 1)
	ack, err := s.processIncomingBleAddress(ctx, in)
	if err != nil {
		log.Println(err)
		return &pb.Reply{Acknowledged: true}, nil
	}
	return &pb.Reply{Acknowledged: *ack}, nil
}

// Addresses Handler for receiving array of IP/MAC requests
func (s *Server) Addresses(ctx context.Context, in *pb.AddressesRequest) (*pb.Reply, error) {
	s.grpcPrometheusMetrics(ctx, "grpc_addresses", "Addresses")
	s.grpcHitsMetrics("grpc_address_count", "Address", len(in.Addresses))
	for _, addr := range in.Addresses {
		_, err := s.ProcessIncomingAddress(ctx, addr)
		if err != nil {
			return nil, err
		}
	}
	return &pb.Reply{Acknowledged: true}, nil
}

// Address Handler for receiving IP/MAC requests
func (s *Server) Address(ctx context.Context, in *pb.AddressRequest) (*pb.Reply, error) {
	s.grpcPrometheusMetrics(ctx, "grpc_address", "Address")
	s.grpcHitsMetrics("grpc_address_count", "Address", 1)
	return s.ProcessIncomingAddress(ctx, in)
}

// ListCommandQueue Handler for Listing all the TimedCommands
func (s *Server) ListCommandQueue(ctx context.Context, _ *empty.Empty) (*pb.CQsResponse, error) {
	//s.grpcPrometheusMetrics(ctx, "grpc_address", "Address")
	//s.grpcHitsMetrics("grpc_address_count", "Address", 1)
	tcs, err := s.getTc()
	if err != nil {
		log.Printf("Error listing CQ: %v", err)
		return nil, err
	}
	cqs := make([]*pb.TimedCommands, 0)
	for _, val := range tcs {
		cq := pb.TimedCommands{
			Id:        val.Id,
			Executeat: val.Executeat,
			Owner:     val.Owner,
			Command:   val.Command,
			Executed:  val.Executed,
		}
		cqs = append(cqs, &cq)
	}
	return &pb.CQsResponse{Cqs: cqs}, nil
}

// ListTimedCommands lists all the TimedCommands basically the bles
func (s *Server) ListTimedCommands(ctx context.Context, _ *empty.Empty) (*pb.TCsResponse, error) {
	//s.grpcPrometheusMetrics(ctx, "grpc_address", "Address")
	//s.grpcHitsMetrics("grpc_address_count", "Address", 1)
	bles, err := s.readBleConfigAsSlice()
	if err != nil {
		log.Printf("Error listing Bles: %v", err)
		return nil, err
	}
	return &pb.TCsResponse{Bles: bles}, nil
}

// ListDevices lists all the Devices
func (s *Server) ListDevices(ctx context.Context, _ *empty.Empty) (*pb.DevicesResponse, error) {
	//s.grpcPrometheusMetrics(ctx, "grpc_address", "Address")
	//s.grpcHitsMetrics("grpc_address_count", "Address", 1)
	devices, err := s.getDevices()
	if err != nil {
		log.Printf("Error listing Devices: %v", err)
		return nil, err
	}
	return &pb.DevicesResponse{Devices: devices}, nil
}

// DeleteCommandQueue Deletes an entire job from CommandQueue
func (s *Server) DeleteCommandQueue(ctx context.Context, request *pb.StringRequest) (*pb.Reply, error) {
	//s.grpcPrometheusMetrics(ctx, "grpc_address", "Address")
	//s.grpcHitsMetrics("grpc_address_count", "Address", 1)
	_, err := s.EtcdClient.Delete(ctx, fmt.Sprintf("%s%s", tcPrefix, request.Key))
	if err != nil {
		return nil, err
	}
	return &pb.Reply{Acknowledged: true}, nil
}

// DeleteTimedCommand Deletes an Individual Timed Command from the CommandQueue
func (s *Server) DeleteTimedCommand(ctx context.Context, request *pb.StringRequest) (*pb.Reply, error) {
	//s.grpcPrometheusMetrics(ctx, "grpc_address", "Address")
	//s.grpcHitsMetrics("grpc_address_count", "Address", 1)
	_, err := s.EtcdClient.Delete(ctx, fmt.Sprintf("%s%s", tcPrefix, request.Key), etcdv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	return &pb.Reply{Acknowledged: true}, nil
}

// CompleteTimedCommands Handler for finishing TimedCommands Now!
func (s *Server) CompleteTimedCommands(ctx context.Context, request *pb.StringRequest) (*pb.Reply, error) {
	//s.grpcPrometheusMetrics(ctx, "grpc_address", "Address")
	//s.grpcHitsMetrics("grpc_address_count", "Address", 1)
	items, err := s.getTc()
	if err != nil {
		return nil, err
	}
	sortedTcs := make([]*pb.TimedCommands, 0)
	for _, tc := range items {
		sortedTcs = append(sortedTcs, tc)
	}
	sort.Sort(ByExecutedAt{sortedTcs})
	for idx, v := range sortedTcs {
		if v.Owner == request.Key {
			v.Executeat = time.Now().Unix() + int64(idx)
		}
		err = s.writeTc(v)
		if err != nil {
			return nil, err
		}
	}
	return &pb.Reply{Acknowledged: true}, nil
}

// CreateTimedCommand Handler for creating TimedCommands!
func (s *Server) CreateTimedCommand(ctx context.Context, request *pb.TimedCommands) (*pb.Reply, error) {
	//s.grpcPrometheusMetrics(ctx, "grpc_address", "Address")
	//s.grpcHitsMetrics("grpc_address_count", "Address", 1)
	err := s.storeTimedCommand(request)
	if err != nil {
		return nil, err
	}

	return &pb.Reply{Acknowledged: true}, nil
}

// CompleteTimedCommand Handler for finishing TimedCommands Now!
func (s *Server) CompleteTimedCommand(ctx context.Context, request *pb.StringRequest) (*pb.Reply, error) {
	//s.grpcPrometheusMetrics(ctx, "grpc_address", "Address")
	//s.grpcHitsMetrics("grpc_address_count", "Address", 1)
	item, err := s.getTcById(request.Key)
	if err != nil {
		return nil, err
	}
	item.Executeat = time.Now().Unix()
	err = s.writeTc(item)
	if err != nil {
		return nil, err
	}

	return &pb.Reply{Acknowledged: true}, nil
}

// DeleteDevice Handler for deleting Devices
func (s *Server) DeleteDevice(ctx context.Context, request *pb.StringRequest) (*pb.Reply, error) {
	//s.grpcPrometheusMetrics(ctx, "grpc_address", "Address")
	//s.grpcHitsMetrics("grpc_address_count", "Address", 1)
	if request.Key == "" {
		return &pb.Reply{Acknowledged: true}, nil
	}
	err := s.deleteDeviceById(request.Key)
	if err != nil {
		return &pb.Reply{Acknowledged: false}, err
	}
	return &pb.Reply{Acknowledged: true}, nil
}

// UpdateDevice Handler for updating Devices
func (s *Server) UpdateDevice(ctx context.Context, request *pb.Devices) (*pb.Reply, error) {
	//s.grpcPrometheusMetrics(ctx, "grpc_address", "Address")
	//s.grpcHitsMetrics("grpc_address_count", "Address", 1)
	err := s.writeNetworkDevice(request)
	if err != nil {
		return &pb.Reply{Acknowledged: false}, err
	}
	return &pb.Reply{Acknowledged: true}, nil
}
