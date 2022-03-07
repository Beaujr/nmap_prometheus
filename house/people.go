package house

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gopkg.in/yaml.v2"
	"log"
	"time"
)

type people struct {
	Id      string   `json:"id",yaml:"id"`
	Devices []string `json:"devices",yaml:"devices"`
	Name    string   `json:"name",yaml:"name"`
}

func (s *Server) createPerson(devices []string, name string) error {
	// Create a queue
	person := &people{
		Name:    name,
		Devices: devices,
		Id:      uuid.New().String(),
	}
	metricId := fmt.Sprintf("%s%s", metricsKey, person.Id)
	if metrics[metricId] == nil {
		metrics[metricId] = promauto.NewGauge(prometheus.GaugeOpts{
			Name: "home_detector_people_devices",
			Help: "BleDevice in home",
			ConstLabels: prometheus.Labels{
				"name": name,
				"id":   person.Id,
			},
		})
	}
	metrics[metricId].Set(float64(time.Now().Unix()))
	err := s.writePerson(person)
	if err != nil {
		log.Println(err)
	}

	return nil
}

func (s *Server) writePerson(person *people) error {
	d1, err := yaml.Marshal(person)
	if err != nil {
		log.Fatalf(err.Error())
	}

	key := fmt.Sprintf("%s%s", peoplePrefix, person.Id)
	_, err = s.EtcdClient.Put(context.Background(), key, string(d1))
	return err
}
