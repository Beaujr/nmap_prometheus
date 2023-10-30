package house

import (
	"context"
	"fmt"
	pb "github.com/beaujr/nmap_prometheus/proto"
	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"
	"log"
	"sort"
	"strings"
	"time"
)

var metricsKey = "cq_"

type tc struct {
	*pb.TimedCommands
}

func (tc tc) observe(ctx context.Context, obs api.Observer) error {
	attrs := []attribute.KeyValue{
		attribute.Key("name").String(strings.ReplaceAll(tc.Id, " ", "_")),
		attribute.Key("command").String(tc.Command),
	}
	val := float64(0)
	if tc.Executeat-int64(time.Now().Unix()) > 0 {
		val = float64(tc.Executeat - int64(time.Now().Unix()))
	}
	obs.ObserveFloat64(cq, val, api.WithAttributes(attrs...))
	return nil
}

func (s *Server) processTimedCommandQueue() error {
	tcs, err := s.getTc()
	if err != nil {
		s.Logger.Error(err.Error())
		return err
	}
	sortedTcs := make([]*pb.TimedCommands, 0)
	for _, tc := range tcs {
		sortedTcs = append(sortedTcs, tc)
	}
	sort.Sort(ByExecutedAt{sortedTcs})
	for _, tc := range tcs {
		err = s.processTimedCommand(tc)
		if err != nil {
			s.Logger.Error(err.Error())
			return err
		}
		if err == nil {
			//metrics["cq"].WithLabelValues(strings.ReplaceAll(tc.Id, " ", "_"), tc.GetCommand()).Set(0)
		}
	}
	return nil
}

func (s *Server) createTimedCommand(timeout int64, id string, commandId string, command string, owner string) error {
	if timeout == 0 {
		go func() {
			log.Printf("Executing immediately: %s", command)
			_, err := s.callAssistant(command)
			log.Printf("Executed")
			if err != nil {
				log.Printf("error calling assistant: %v", err)
				log.Printf("Creating TC instead")
				err = s.createTimedCommand(1, id, commandId, command, owner)
				if err != nil {
					log.Printf("failed to schedule action: %v", err)
				}
			}
		}()
		return nil
	}
	// Create a tc
	tc := &pb.TimedCommands{
		Owner:     owner,
		Command:   command,
		Executeat: int64(time.Now().Unix()) + timeout,
		Executed:  false,
		Id:        fmt.Sprintf("%s%v", id, commandId),
	}
	return s.storeTimedCommand(tc)
}

func (s *Server) storeTimedCommand(timedCommand *pb.TimedCommands) error {
	_, err := meter.RegisterCallback(tc{timedCommand}.observe, cq)
	if err != nil {
		log.Panicln(err.Error())
	}
	err = s.writeTc(timedCommand)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) processTimedCommand(tc *pb.TimedCommands) error {
	if tc.Executeat < int64(time.Now().Unix()) && !tc.Executed && *cqEnabled {
		_, err := s.callAssistant(tc.Command)
		if err != nil {
			s.Logger.Error(err.Error())
			return err
		}
		err = s.deleteTc(tc)
		if err != nil {
			s.Logger.Error(err.Error())
		}
		//err = s.writeTc(tc)
		//if err != nil {
		//	s.Logger.Error(err)
		//}
		err = s.NotificationClient.SendNotification("Scheduled Task", tc.Command, "devices")
		if err != nil {
			s.Logger.Error(err.Error())
		}

	}
	return nil
}
