package house

import (
	"fmt"
	pb "github.com/beaujr/nmap_prometheus/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"log"
	"sort"
	"strings"
	"time"
)

var metricsKey = "cq_"

func (s *Server) processTimedCommandQueue() error {
	tcs, err := s.getTc()
	if err != nil {
		log.Println(err)
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
			log.Println(err)
			//tc.Executed = false
			//err = s.writeTc(tc)
			return err
		}
	}

	if len(tcs) == 0 {
		for key, val := range metrics {
			if strings.Contains(key, metricsKey) && val != nil {
				metrics[key].Set(0)
			}
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

func (s *Server) storeTimedCommand(tc *pb.TimedCommands) error {
	metricId := fmt.Sprintf("%s%s", metricsKey, tc.Id)
	if metrics[metricId] == nil {
		metrics[metricId] = promauto.NewGauge(prometheus.GaugeOpts{
			Name: "home_detector_ble_device",
			Help: "BleDevice in home",
			ConstLabels: prometheus.Labels{
				"name":    strings.ReplaceAll(tc.Id, " ", "_"),
				"command": tc.Command,
			},
		})
	}
	metrics[metricId].Set(float64(tc.Executeat))

	err := s.writeTc(tc)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) processTimedCommand(tc *pb.TimedCommands) error {
	metricId := fmt.Sprintf("%s%s", metricsKey, tc.Id)
	if metrics[metricId] == nil {
		metrics[metricId] = promauto.NewGauge(prometheus.GaugeOpts{
			Name: "home_detector_ble_device",
			Help: "BleDevice in home",
			ConstLabels: prometheus.Labels{
				"name":    strings.ReplaceAll(tc.Id, " ", "_"),
				"command": tc.Command,
			},
		})
	}
	if tc.Executeat-int64(time.Now().Unix()) > 0 {
		metrics[metricId].Set(float64(tc.Executeat - int64(time.Now().Unix())))
	} else if tc.Executeat-int64(time.Now().Unix()) < 0 {
		metrics[metricId].Set(float64(0))
	}
	if tc.Executeat < int64(time.Now().Unix()) && !tc.Executed && *cqEnabled {
		_, err := s.callAssistant(tc.Command)
		if err != nil {
			log.Println(err)
			return err
		}
		err = s.deleteTc(tc)
		if err != nil {
			log.Println(err)
		}
		//err = s.writeTc(tc)
		//if err != nil {
		//	log.Println(err)
		//}
		err = s.notificationClient.SendNotification("Scheduled Task", tc.Command, "devices")
		if err != nil {
			log.Println(err)
		}

	}
	return nil
}
