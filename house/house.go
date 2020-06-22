package house

import (
	"context"
	"flag"
	"fmt"
	"github.com/beaujr/nmap_prometheus/assistant"
	"github.com/beaujr/nmap_prometheus/notifications"
	pb "github.com/beaujr/nmap_prometheus/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type networkId struct {
	Ip  string `json:"ip",yaml:"ip"`
	Mac string `json:"mac",yaml:"mac"`
}

type device struct {
	Id       networkId `json:"id",yaml:"id"`
	LastSeen int64     `json:"lastSeen",yaml:"lastSeen"`
	Away     bool      `json:"away",yaml:"away"`
	Name     string    `json:"name",yaml:"name"`
	Person   bool      `json:"person",yaml:"person"`
	Command  string    `json:"command",yaml:"command"`
	Smart    bool      `json:"smart",yaml:"smart"`
}

//IOT iotDevices
var timeAwaySeconds = flag.Int64("timeout", 300, "")
var houseDevices = []*device{}
var iotDevices = []*device{}
var syncStatusWithGA = time.Hour.Seconds()
var metrics map[string]prometheus.Gauge
var debug = flag.Bool("debug", false, "Debug mode")

var (
	peopleHome = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "home_detector_people_home",
		Help: "The total number of houseDevices at home",
	})
)

// HomeManager manages devices and metric collection
type HomeManager interface {
	adjustLights(lightGroup string, brightness string) error
	deviceDetectState(phone device) int64
	deviceManager() error
	isDeviceOn(iot *device) (bool, error)
	isHouseEmpty() bool
	httpHealthCheck(url string) bool
	iotdeviceManager(iotDevice *device) error
	iotStatusManager() error
	recordMetrics()
}

// Server is an implementation of the proto HomeDetectorServer
type Server struct {
	pb.UnimplementedHomeDetectorServer
}

func readConfig(filename string) ([]*device, error) {
	// Open our yamlFile
	yamlFile, err := os.Open(filename)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	log.Println(fmt.Sprintf("Successfully Opened: %s", filename))
	defer yamlFile.Close()

	byteValue, err := ioutil.ReadAll(yamlFile)
	if err != nil {
		return nil, err
	}

	var result []*device
	err = yaml.Unmarshal(byteValue, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func writeConfig(filename string) error {
	d1, err := yaml.Marshal(append(houseDevices, iotDevices...))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, d1, 0644)
	if err != nil {
		return err
	}
	return nil
}

// NewServer new instance of HomeManager
func NewServer() HomeManager {
	server := &Server{}
	allDevices, err := readConfig("config/devices.yaml")
	if err != nil {
		log.Println(err)
	}
	for _, item := range allDevices {
		if item.Smart {
			iotDevices = append(iotDevices, item)
			continue
		}
		houseDevices = append(houseDevices, item)
	}

	c := cron.New(cron.WithSeconds())
	c.AddFunc("*/10 * * * * *", func() {
		err := server.deviceManager()
		if err != nil {
			log.Println(err)
		}
	})

	c.AddFunc("*/10 * * * * *", func() {
		err := server.iotStatusManager()
		if err != nil {
			log.Println(err)
		}
	})

	metrics = make(map[string]prometheus.Gauge)
	for _, item := range append(iotDevices, houseDevices...) {
		log.Println(fmt.Sprintf("Registering metric for %s", item.Name))
		server.registerMetric(*item)
	}

	c.Start()
	server.recordMetrics()
	return server
}

func (s *Server) registerMetric(item device) {
	if metrics[item.Name] == nil {
		metrics[item.Name] = promauto.NewGauge(prometheus.GaugeOpts{
			Name: "home_detector_device",
			Help: "Device in home",
			ConstLabels: prometheus.Labels{
				"name": strings.ReplaceAll(item.Name, " ", "_"),
				"mac":  item.Id.Mac,
				"ip":   item.Id.Ip,
			},
		})
		metrics[item.Name+"_lastseen"] = promauto.NewGauge(prometheus.GaugeOpts{
			Name: "home_detector_device_lastseen",
			Help: "Device in home last seen",
			ConstLabels: prometheus.Labels{
				"name": strings.ReplaceAll(item.Name, " ", "_"),
				"mac":  item.Id.Mac,
				"ip":   item.Id.Ip,
			},
		})
	}
}

func (s *Server) recordMetrics() {
	go func() {
		for {
			homeCounter := 0
			for _, person := range houseDevices {
				if !person.Away && person.Person {
					homeCounter++
				}
			}
			peopleHome.Set(float64(homeCounter))

			for _, item := range append(iotDevices, houseDevices...) {
				state := 1
				if item.Away {
					state = 0
				}
				if metrics[item.Name] != nil {
					metrics[item.Name].Set(float64(state))
				}
				if metrics[item.Name+"_lastseen"] != nil {
					metrics[item.Name+"_lastseen"].Set(float64(item.LastSeen))
				}
			}
			time.Sleep(2 * time.Second)
		}
	}()
}
func (s *Server) adjustLights(lightGroup string, brightness string) error {
	_, err := s.callAssistant("set " + lightGroup + " lights to " + brightness + "% brightness")
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) callAssistant(command string) (*string, error) {
	if *debug {
		return &command, nil
	}
	return assistant.Call(command)
}

func (s *Server) deviceDetectState(item device) int64 {
	lastSeen := int64(time.Now().Unix()) - item.LastSeen
	return lastSeen
}

// Address Handler for receiving IP/MAC requests
func (s *Server) Address(ctx context.Context, in *pb.AddressRequest) (*pb.Reply, error) {
	incoming := in
	newDevice := true
	for _, houseDevice := range houseDevices {
		if houseDevice.Id.Ip == incoming.Ip &&
			houseDevice.Id.Mac == incoming.Mac ||
			houseDevice.Id.Mac == incoming.Mac && incoming.Mac != "" ||
			houseDevice.Id.Ip == incoming.Ip && incoming.Mac == "" ||
			houseDevice.Id.Ip != incoming.Ip && incoming.Mac != "" && incoming.Mac == houseDevice.Id.Mac ||
			houseDevice.Id.Ip == incoming.Ip && incoming.Ip != "" && incoming.Mac != houseDevice.Id.Mac {

			if incoming.Ip != "" {
				houseDevice.Id.Ip = incoming.Ip
			} else if incoming.Mac != "" {
				houseDevice.Id.Mac = incoming.Mac
			}
			newDevice = false
			log.Println(houseDevice.Name)
			timeAway := s.deviceDetectState(*houseDevice)
			if timeAway > *timeAwaySeconds && houseDevice.Person {
				log.Println(fmt.Sprintf("Device: %s has returned after %d seconds", houseDevice.Name, timeAway))
				if *debug {
					log.Printf("Notification: %s, %s", houseDevice.Name, "is home")
				} else {
					err := notifications.SendNotification(houseDevice.Name, "has returned home.")
					if err != nil {
						return nil, err
					}
				}

			}
			houseDevice.Away = false
			houseDevice.LastSeen = int64(time.Now().Unix())
		}
	}
	if newDevice {
		name := in.Ip
		if in.Mac != "" {
			name = in.Mac
		}
		newDevice := device{
			Name:     strings.ReplaceAll(strings.ReplaceAll(name, ".", "_"), ":", "_"),
			Id:       networkId{Ip: in.Ip, Mac: in.Mac},
			Away:     false,
			LastSeen: int64(time.Now().Unix()),
			Person:   false,
			Command:  "",
		}

		log.Println(fmt.Printf("New Device: %s", name))
		if !*debug {
			err := notifications.SendNotification(newDevice.Name, "New Device")
			if err != nil {
				return nil, err
			}
		}

		houseDevices = append(houseDevices, &newDevice)
		err := writeConfig("config/devices.yaml")
		if err != nil {
			return nil, err
		}
		s.registerMetric(newDevice)
	}
	return &pb.Reply{Acknowledged: true}, nil
}

func (s *Server) httpHealthCheck(url string) bool {
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("cache-control", "no-cache")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	return res.StatusCode == 200
}

func (s *Server) isDeviceOn(iot *device) (bool, error) {
	lastSeen := s.deviceDetectState(*iot)
	if lastSeen > int64(syncStatusWithGA) {
		state, err := s.callAssistant(iot.Command)
		if err != nil {
			return false, nil
		}
		iot.LastSeen = int64(time.Now().Unix())
		iot.Away = *state == "off"
	}

	return !iot.Away, nil
}

func (s *Server) isHouseEmpty() bool {
	houseEmpty := true
	for _, device := range houseDevices {
		if !device.Away && device.Person {
			houseEmpty = false
		}
	}
	return houseEmpty
}

func (s *Server) iotStatusManager() error {
	for _, device := range iotDevices {
		err := s.iotdeviceManager(device)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) iotdeviceManager(iotDevice *device) error {
	houseEmpty := s.isHouseEmpty()
	on, err := s.isDeviceOn(iotDevice)
	if err != nil {
		return err
	}
	if houseEmpty && on {
		command := fmt.Sprintf("Turning %s off", iotDevice.Name)
		log.Println(command)
		if !*debug {
			err := notifications.SendNotification("House Empty", command)
			if err != nil {
				return nil
			}
		}
		_, err := s.callAssistant(fmt.Sprintf("turn %s off", iotDevice.Name))
		if err != nil {
			return err
		}
		iotDevice.Away = true
	}
	return nil
}

func (s *Server) deviceManager() error {
	for _, device := range houseDevices {
		timeAway := s.deviceDetectState(*device)
		if timeAway > *timeAwaySeconds && !device.Away {
			log.Println(fmt.Sprintf("Device: %s has left after %d seconds", device.Name, timeAway))
			device.Away = true
			if device.Person {
				err := notifications.SendNotification(device.Name, "Has left the house")
				if err != nil {
					return nil
				}
			}
		}
	}
	return nil
}
