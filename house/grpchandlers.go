package house

import (
	"context"
	"fmt"
	pb "github.com/beaujr/nmap_prometheus/proto"
	"github.com/golang/protobuf/ptypes/empty"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Ack for bluetooth reported MAC addresses
func (s *Server) Ack(ctx context.Context, in *pb.BleRequest) (*pb.Reply, error) {
	s.grpcPrometheusMetrics(ctx, "grpc_ble", "Ack")
	s.grpcHitsMetrics(ctx, "Ack", 1)
	ack, err := s.processIncomingBleAddress(ctx, in)
	if err != nil {
		s.Logger.Error(err.Error())
		return &pb.Reply{Acknowledged: true}, nil
	}
	return &pb.Reply{Acknowledged: *ack}, nil
}

// Addresses Handler for receiving array of IP/MAC requests
func (s *Server) Addresses(ctx context.Context, in *pb.AddressesRequest) (*pb.Reply, error) {
	s.grpcPrometheusMetrics(ctx, "grpc_addresses", "Addresses")
	s.grpcHitsMetrics(ctx, "Address", len(in.Addresses))
	var wg sync.WaitGroup
	errChan := make(chan error, len(in.Addresses))
	for _, addr := range in.Addresses {
		wg.Add(1)
		go func(ctx *context.Context, address *pb.AddressRequest) {
			defer wg.Done()
			_, err := s.ProcessIncomingAddress(*ctx, address)
			errChan <- err
		}(&ctx, addr)
	}
	go func() {
		wg.Wait()
		close(errChan)
	}()
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	return &pb.Reply{Acknowledged: true}, nil
}

// Address Handler for receiving IP/MAC requests
func (s *Server) Address(ctx context.Context, in *pb.AddressRequest) (*pb.Reply, error) {
	s.grpcPrometheusMetrics(ctx, "grpc_address", "Address")
	s.grpcHitsMetrics(ctx, "Address", 1)
	return s.ProcessIncomingAddress(ctx, in)
}

func (s *Server) ListPeople(ctx context.Context, _ *emptypb.Empty) (*pb.PeopleResponse, error) {
	return s.ListPeopleRequest(ctx)
}

func (s *Server) TogglePerson(ctx context.Context, device *pb.Devices) (*pb.Reply, error) {
	initialEmptyState := s.IsHouseEmpty(ctx, device.GetHome())
	// im only toggling person atm best to assume this is the only field changing for now
	// this will be the old key path
	path := "device"
	if device.GetPerson() {
		path = "person"
	}

	_, err := s.UpdateDevice(ctx, device)
	if err != nil {
		return nil, err
	}
	if !device.Away {
		log.Printf("Toggling %s to alive value %s\n", device.GetId().GetMac(), path)
		_, err = s.Kv.Put(ctx, filepath.Join(AlivePrefix, device.GetHome(), device.GetId().GetMac()), path)
		if err != nil {
			return nil, err
		}
	}

	currentEmptyState := s.IsHouseEmpty(ctx, device.GetHome())
	if currentEmptyState != initialEmptyState {
		err = s.ToggleHouseStatus(device.GetHome(), currentEmptyState)
		if err != nil {
			return &pb.Reply{Acknowledged: true}, err
		}
	}
	return &pb.Reply{Acknowledged: true}, err
}

// ListCommandQueue Handler for Listing all the TimedCommands
func (s *Server) ListCommandQueue(ctx context.Context, _ *empty.Empty) (*pb.CQsResponse, error) {
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
	//s.GrpcPrometheusMetrics(ctx, "grpc_address", "Address")
	//s.GrpcHitsMetrics("grpc_address_count", "Address", 1)
	bles, err := s.readBleConfigAsSlice()
	if err != nil {
		log.Printf("Error listing Bles: %v", err)
		return nil, err
	}
	return &pb.TCsResponse{Bles: bles}, nil
}

// ListDevices lists all the Devices
func (s *Server) ListDevices(ctx context.Context, _ *empty.Empty) (*pb.DevicesResponse, error) {
	//s.GrpcPrometheusMetrics(ctx, "grpc_address", "Address")
	//s.GrpcHitsMetrics("grpc_address_count", "Address", 1)
	devices, err := s.getDevices(ctx)
	if err != nil {
		log.Printf("Error listing Devices: %v", err)
		return nil, err
	}
	return &pb.DevicesResponse{Devices: devices}, nil
}

// DeleteCommandQueue Deletes an entire job from CommandQueue
func (s *Server) DeleteCommandQueue(ctx context.Context, request *pb.StringRequest) (*pb.Reply, error) {
	//s.GrpcPrometheusMetrics(ctx, "grpc_address", "Address")
	//s.GrpcHitsMetrics("grpc_address_count", "Address", 1)
	_, err := s.Kv.Delete(ctx, fmt.Sprintf("%s%s", tcPrefix, request.Key))
	if err != nil {
		return nil, err
	}
	return &pb.Reply{Acknowledged: true}, nil
}

// DeleteTimedCommand Deletes an Individual Timed Command from the CommandQueue
func (s *Server) DeleteTimedCommand(ctx context.Context, request *pb.StringRequest) (*pb.Reply, error) {
	//s.GrpcPrometheusMetrics(ctx, "grpc_address", "Address")
	//s.GrpcHitsMetrics("grpc_address_count", "Address", 1)
	_, err := s.Kv.Delete(ctx, fmt.Sprintf("%s%s", tcPrefix, request.Key), etcdv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	return &pb.Reply{Acknowledged: true}, nil
}

// CompleteTimedCommands Handler for finishing TimedCommands Now!
func (s *Server) CompleteTimedCommands(ctx context.Context, request *pb.StringRequest) (*pb.Reply, error) {
	//s.GrpcPrometheusMetrics(ctx, "grpc_address", "Address")
	//s.GrpcHitsMetrics("grpc_address_count", "Address", 1)
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
	//s.GrpcPrometheusMetrics(ctx, "grpc_address", "Address")
	//s.GrpcHitsMetrics("grpc_address_count", "Address", 1)
	err := s.storeTimedCommand(request)
	if err != nil {
		return nil, err
	}

	return &pb.Reply{Acknowledged: true}, nil
}

// CompleteTimedCommand Handler for finishing TimedCommands Now!
func (s *Server) CompleteTimedCommand(ctx context.Context, request *pb.StringRequest) (*pb.Reply, error) {
	//s.GrpcPrometheusMetrics(ctx, "grpc_address", "Address")
	//s.GrpcHitsMetrics("grpc_address_count", "Address", 1)
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
	//s.GrpcPrometheusMetrics(ctx, "grpc_address", "Address")
	//s.GrpcHitsMetrics("grpc_address_count", "Address", 1)
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
	//s.GrpcPrometheusMetrics(ctx, "grpc_address", "Address")
	//s.GrpcHitsMetrics("grpc_address_count", "Address", 1)
	err := s.WriteNetworkDevice(ctx, request)
	if err != nil {
		return &pb.Reply{Acknowledged: false}, err
	}
	return &pb.Reply{Acknowledged: true}, nil
}

// UpdateDevice Handler for updating Devices
func (s *Server) HouseEmpty(ctx context.Context, request *pb.StringRequest) (*pb.Reply, error) {
	empty := s.IsHouseEmpty(ctx, request.Key)
	return &pb.Reply{Acknowledged: empty}, nil
}
