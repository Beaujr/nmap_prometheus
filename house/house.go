package house

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/beaujr/nmap_prometheus/assistant"
	"github.com/beaujr/nmap_prometheus/notifications"
	pb "github.com/beaujr/nmap_prometheus/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type networkId struct {
	Ip   string `json:"ip",yaml:"ip"`
	Mac  string `json:"mac",yaml:"mac"`
	UUID string `json:"uuid",yaml:"uuid"`
}

//IOT iotDevices
var timeAwaySeconds = flag.Int64("timeout", 300, "")
var networkConfigFile = flag.String("config", "config/devices.yaml", "Path to config file")
var bleConfigFile = flag.String("bleconfig", "config/ble_devices.yaml", "Path to config file")
var houseDevices = []*device{}
var iotDevices = []*device{}
var bleDevices = []*bleDevice{}
var commandQueue = []*TimedCommand{}
var syncStatusWithGA = time.Hour.Seconds()
var metrics map[string]prometheus.Gauge
var gHouseEmpty = false
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
	deviceDetectState(phone int64) int64
	deviceManager() error
	isDeviceOn(iot *device) (bool, error)
	isHouseEmpty() bool
	httpHealthCheck(url string) bool
	iotdeviceManager(iotDevice *device, empty bool) error
	iotStatusManager() error
	recordMetrics()
	Devices(w http.ResponseWriter, req *http.Request)
	People(w http.ResponseWriter, req *http.Request)
}

// Server is an implementation of the proto HomeDetectorServer
type Server struct {
	pb.UnimplementedHomeDetectorServer
}

func writeConfig(data []byte, filename string) error {
	err := ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

// NewServer new instance of HomeManager
func NewServer() HomeManager {
	server := &Server{}
	cfgDevices, err := readNetworkConfig(*networkConfigFile)
	if err != nil {
		log.Println(err)
	}
	allDevices, err := uniqueNetwork(cfgDevices)
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

	// Bluetooth
	savedBle, err := readBleConfig(*bleConfigFile)
	bleDevices = savedBle
	if err != nil {
		log.Println(err)
	}
	//bleDevices, err = uniqueBle(cfgBleDevices)
	//if err != nil {
	//	log.Println(err)
	//}

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
	//commandQueue = append(commandQueue, &TimedCommand{
	//	Command:   "turn christmas tree off",
	//	ExecuteAt: 0,
	//})
	c.AddFunc("*/10 * * * * *", func() {
		for _, tc := range commandQueue {
			if tc.ExecuteAt < int64(time.Now().Unix()) && !tc.Executed {
				tc.Executed = true
				if !*debug {
					_, err := server.callAssistant(tc.Command)
					if err != nil {
						log.Println(err)
					}
					err = notifications.SendNotification("Scheduled Task", tc.Command)
					if err != nil {
						log.Println(err)
					}
				} else {
					log.Printf("Scheduled Task: %s", tc.Command)
				}
			}
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

// Devices API endpoint to determine devices status
func (s *Server) Devices(w http.ResponseWriter, req *http.Request) {
	js, err := json.Marshal(append(iotDevices, houseDevices...))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	return
}

// People API endpoint to determine person device status
func (s *Server) People(w http.ResponseWriter, req *http.Request) {
	people := make([]*device, 0)
	for _, device := range houseDevices {
		if device.Person {
			people = append(people, device)
		}
	}
	js, err := json.Marshal(people)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	return
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

func (s *Server) deviceDetectState(deviceLastSeen int64) int64 {
	now := int64(time.Now().Unix())
	lastSeen := now - deviceLastSeen
	return lastSeen
}

func (s *Server) newBleDevice(in *pb.BleRequest) error {
	newDevice := bleDevice{
		Id:       in.Mac,
		LastSeen: int64(time.Now().Unix()),
		Commands: make([]command, 0),
	}

	log.Println(fmt.Printf("New BLE Device: %s", in.Mac))

	bleDevices = append(bleDevices, &newDevice)
	_, err := uniqueBle(bleDevices)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) newDevice(in *pb.AddressRequest) error {
	name := in.Ip
	if in.Mac != "" {
		name = in.Mac
	}
	newDevice := device{
		Name:     strings.ReplaceAll(strings.ReplaceAll(name, ".", "_"), ":", "_"),
		Id:       networkId{Ip: in.Ip, Mac: in.Mac, UUID: name},
		Away:     false,
		LastSeen: int64(time.Now().Unix()),
		Person:   false,
		Command:  "",
		Home:     in.Home,
	}

	log.Println(fmt.Printf("New Device: %s", name))
	if !*debug {
		err := notifications.SendNotification(newDevice.Name, fmt.Sprintf("New Device in %", newDevice.Home))
		if err != nil {
			return err
		}
	}

	houseDevices = append(houseDevices, &newDevice)
	_, err := uniqueNetwork(houseDevices)
	if err != nil {
		return err
	}
	s.registerMetric(newDevice)
	return nil
}

func (s *Server) existingDevice(houseDevice *device, incoming *pb.AddressRequest) error {
	if incoming.Ip != "" && incoming.Ip != houseDevice.Id.Ip {
		houseDevice.Id.Ip = incoming.Ip
		err := writeNetworkDevices(houseDevices)
		if err != nil {
			log.Printf("Error updating: %s", houseDevice.Id.UUID)
		}
	}

	if incoming.Mac != "" {
		houseDevice.Id.Mac = incoming.Mac
	}

	if incoming.Mac == "" && incoming.Ip == houseDevice.Id.UUID {
		houseDevice.Id.Ip = incoming.Ip
	}

	//log.Println(houseDevice.Name)
	timeAway := s.deviceDetectState(houseDevice.LastSeen)
	if timeAway > *timeAwaySeconds && houseDevice.Person {
		log.Println(fmt.Sprintf("Device: %s has returned after %d seconds", houseDevice.Name, timeAway))
		if *debug {
			log.Printf("Notification: %s, %s", houseDevice.Name, "is home")
		} else {
			err := notifications.SendNotification(houseDevice.Name, "has returned home.")
			if err != nil {
				return err
			}
		}
	}
	houseDevice.Away = false
	houseDevice.LastSeen = int64(time.Now().Unix())
	return nil
}

// Address Handler for receiving IP/MAC requests
func (s *Server) Address(ctx context.Context, in *pb.AddressRequest) (*pb.Reply, error) {
	incoming := in
	newDevice := true
	for _, houseDevice := range houseDevices {
		if (houseDevice.Id.UUID == incoming.Mac ||
			houseDevice.Id.UUID == incoming.Ip) && incoming.Home == houseDevice.Home {
			//if houseDevice.Id.Ip == incoming.Ip && houseDevice.Id.Mac == incoming.Mac ||
			//	houseDevice.Id.Mac == incoming.Mac && incoming.Mac != "" ||
			//	houseDevice.Id.Ip == incoming.Ip && incoming.Mac == "" ||
			//	houseDevice.Id.Ip != incoming.Ip && incoming.Mac != "" && incoming.Mac == houseDevice.Id.Mac ||
			//	houseDevice.Id.Ip == incoming.Ip && incoming.Ip != "" && incoming.Mac != houseDevice.Id.Mac {
			newDevice = false
			err := s.existingDevice(houseDevice, incoming)
			if err != nil {
				return nil, err
			}
			s.iotStatusManager()

		}
	}

	if newDevice {
		err := s.newDevice(in)
		if err != nil {
			return nil, err
		}
	}

	return &pb.Reply{Acknowledged: true}, nil
}

// Ack for bluetooth reported MAC addresses
func (s *Server) Ack(ctx context.Context, in *pb.BleRequest) (*pb.Reply, error) {
	newDevice := true
	device := &bleDevice{}
	for _, houseDevice := range bleDevices {
		if in.Mac == houseDevice.Id {
			newDevice = false
			device = houseDevice
		}
	}
	if !newDevice {
		lastSeen := s.deviceDetectState(device.LastSeen)
		device.LastSeen = int64(time.Now().Unix())
		err := writeBleDevices(bleDevices)
		if err != nil {
			log.Printf("Error updating BLE device: %s", device.Id)
		}
		if lastSeen > *timeAwaySeconds {
			log.Printf("BLE: %s (%s) detected", device.Name, device.Id)
			for _, command := range device.Commands {
				if command.Timeout > 0 {
					// Create a queue
					commandQueue = append(commandQueue, &TimedCommand{
						Command:   command.TimeoutCommand,
						ExecuteAt: int64(time.Now().Unix()) + command.Timeout,
						Executed:  false,
					})
					fmt.Println(&commandQueue)
				}
				if !*debug {
					_, err := s.callAssistant(command.Command)
					if err != nil {
						log.Println(err)
					}
					err = notifications.SendNotification(device.Name, command.Command)
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
	}
	return &pb.Reply{Acknowledged: !newDevice}, nil
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
	lastSeen := s.deviceDetectState(iot.LastSeen)
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
	houseEmpty := s.isHouseEmpty()
	if gHouseEmpty == houseEmpty {
		return nil
	}
	for _, device := range iotDevices {
		err := s.iotdeviceManager(device, gHouseEmpty)
		if err != nil {
			return err
		}

	}
	return nil
}

func (s *Server) returnToHouseManager(iotDevice *device) error {
	houseEmpty := s.isHouseEmpty()
	if gHouseEmpty != houseEmpty {
		// the house isnt empty but it was until this device returned
		if !houseEmpty {
			gHouseEmpty = houseEmpty
			command := fmt.Sprintf("%s returned home", iotDevice.Name)
			err := notifications.SendNotification("House no longer Empty", command)
			if err != nil {
				return nil
			}
		}
	}
	return nil
}

func (s *Server) iotdeviceManager(iotDevice *device, houseEmpty bool) error {
	if houseEmpty && iotDevice.PresenceAware {
		gHouseEmpty = houseEmpty
		command := fmt.Sprintf("Turning %s off", iotDevice.Name)
		log.Println(command)
		if !*debug {
			err := notifications.SendNotification("House Empty", command)
			if err != nil {
				return nil
			}
			_, err = s.callAssistant(fmt.Sprintf("turn %s off", iotDevice.Name))
			if err != nil {
				return err
			}
		}
		iotDevice.Away = true
	}
	return nil
}

func (s *Server) deviceManager() error {
	for _, device := range houseDevices {
		//log.Println(device.Name)
		timeAway := s.deviceDetectState(device.LastSeen)
		if timeAway > *timeAwaySeconds && !device.Away {
			log.Println(fmt.Sprintf("Device: %s has left after %d seconds", device.Name, timeAway))
			device.Away = true
			if device.Person {
				err := s.iotStatusManager()
				if err != nil {
					return nil
				}
				if !*debug {
					err := notifications.SendNotification(device.Name, "Has left the house")
					if err != nil {
						return nil
					}
				} else {
					log.Printf("Notification: %s: %s", device.Name, "Has left the house")
				}

			}
		}
	}
	return nil
}
