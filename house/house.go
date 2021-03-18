package house

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/beaujr/nmap_prometheus/assistant"
	"github.com/beaujr/nmap_prometheus/etcd"
	"github.com/beaujr/nmap_prometheus/macvendor"
	"github.com/beaujr/nmap_prometheus/notifications"
	pb "github.com/beaujr/nmap_prometheus/proto"
	etcdv3 "github.com/ozonru/etcd/v3/clientv3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/robfig/cron/v3"
	"google.golang.org/grpc/peer"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
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
var etcdServers = flag.String("etcdServers", "192.168.1.112:2379", "Comma Separated list of etcd servers")
var debug = flag.Bool("debug", false, "Debug mode")

var bleDevices = []*bleDevice{}
var commandQueue = []*TimedCommand{}
var syncStatusWithGA = time.Hour.Seconds()
var metrics map[string]prometheus.Gauge

//var gHouseEmpty map[string]bool

var devicesPrefix = "/devices/"
var homePrefix = "/homes/"
var blesPrefix = "/bles"

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
	isHouseEmpty(home string) bool
	httpHealthCheck(url string) bool
	iotStatusManager() error
	recordMetrics()
	Devices(w http.ResponseWriter, req *http.Request)
	People(w http.ResponseWriter, req *http.Request)
	HomeEmptyState(w http.ResponseWriter, req *http.Request)
}

// Server is an implementation of the proto HomeDetectorServer
type Server struct {
	pb.UnimplementedHomeDetectorServer
	etcdClient etcdv3.KV
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
	etcdClient := etcd.NewClient([]string{*etcdServers})

	server := &Server{etcdClient: etcdClient}
	_, err := server.readNetworkConfig()
	if err != nil {
		log.Println(err)
	}
	// Bluetooth
	savedBle, err := readBleConfig(*bleConfigFile)
	bleDevices = savedBle
	if err != nil {
		log.Println(err)
	}

	c := cron.New(cron.WithSeconds())
	c.AddFunc("*/10 * * * * *", func() {
		err := server.deviceManager()
		if err != nil {
			log.Println(err)
		}
	})
	c.AddFunc("* */1 * * * *", func() {
		knowDevices, err := server.readNetworkConfig()
		if err != nil {
			log.Printf(err.Error())
		}
		for _, val := range knowDevices {
			server.registerMetric(*val)
		}
	})

	c.AddFunc("*/10 * * * * *", func() {
		for _, tc := range commandQueue {
			if tc.ExecuteAt < int64(time.Now().Unix()) && !tc.Executed {
				tc.Executed = true
				if !*debug {
					_, err := server.callAssistant(tc.Command)
					if err != nil {
						log.Println(err)
					}
					err = notifications.SendNotification("Scheduled Task", tc.Command, "devices")
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

	knowDevices, err := server.readNetworkConfig()
	if err != nil {
		log.Printf(err.Error())
	}
	homes := make([]string, 0)
	for _, item := range knowDevices {
		log.Println(fmt.Sprintf("Registering metric for %s", item.Name))
		server.registerMetric(*item)
		homes = append(homes, item.Home)
	}

	for _, home := range homes {
		homeKey := fmt.Sprintf("%s%s", homePrefix, home)
		_, err := server.etcdClient.Put(context.Background(), homeKey, strconv.FormatBool(server.isHouseEmpty(home)))
		if err != nil {
			log.Panic(err.Error())
		}
	}

	c.Start()
	server.recordMetrics()
	return server
}

// Devices API endpoint to determine devices status
func (s *Server) Devices(w http.ResponseWriter, req *http.Request) {
	devices := make([]*device, 0)
	items, err := s.readNetworkConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, device := range items {
		devices = append(devices, device)
	}
	js, err := json.Marshal(devices)
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
	devices, err := s.readNetworkConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, device := range devices {
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

// HomeEmptyState API endpoint to determine house empty status
func (s *Server) HomeEmptyState(w http.ResponseWriter, req *http.Request) {
	homes, err := s.readHomesConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	js, err := json.Marshal(homes)
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
				"name":   strings.ReplaceAll(item.Name, " ", "_"),
				"mac":    item.Id.Mac,
				"ip":     item.Id.Ip,
				"home":   item.Home,
				"person": strconv.FormatBool(item.Person),
			},
		})
		metrics[item.Name+"_lastseen"] = promauto.NewGauge(prometheus.GaugeOpts{
			Name: "home_detector_device_lastseen",
			Help: "Device in home last seen",
			ConstLabels: prometheus.Labels{
				"name": strings.ReplaceAll(item.Name, " ", "_"),
				"mac":  item.Id.Mac,
				"ip":   item.Id.Ip,
				"home": item.Home,
			},
		})
	}
}

func (s *Server) recordMetrics() {
	go func() {
		for {
			knowDevices, err := s.readNetworkConfig()
			if err != nil {
				log.Printf(err.Error())
			}
			homeCounter := 0
			for _, person := range knowDevices {
				if !person.Away && person.Person {
					homeCounter++
				}
			}
			peopleHome.Set(float64(homeCounter))

			for _, item := range knowDevices {
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
	vendor := "unknown"
	if in.Mac != in.Ip && strings.Contains(in.Mac, ":") {
		macVendor, err := macvendor.GetManufacturer(in.Mac)
		if macVendor != nil {
			vendor = *macVendor
		}
		if err != nil {
			log.Printf(err.Error())
			vendor = name
		}

	}

	newDevice := device{
		Name:         strings.ReplaceAll(strings.ReplaceAll(name, ".", "_"), ":", "_"),
		Id:           networkId{Ip: in.Ip, Mac: in.Mac, UUID: name},
		Away:         false,
		LastSeen:     int64(time.Now().Unix()),
		Person:       false,
		Command:      "",
		Manufacturer: vendor,
		Home:         in.Home,
	}

	log.Println(fmt.Printf("New Device: %s", name))
	if !*debug {
		err := notifications.SendNotification(fmt.Sprintf("New Device in %s (%s)", newDevice.Home, newDevice.Id.Ip), newDevice.Manufacturer, newDevice.Home)
		if err != nil {
			return err
		}
	}

	err := s.writeNetworkDevice(&newDevice)
	if err != nil {
		log.Printf("Error saving to ETCD: %s", err.Error())
	}
	s.registerMetric(newDevice)
	return nil
}

func (s *Server) existingDevice(houseDevice *device, incoming *pb.AddressRequest) error {

	if incoming.Mac != "" {
		houseDevice.Id.Mac = incoming.Mac
	}

	if incoming.Mac == "" && incoming.Ip == houseDevice.Id.UUID {
		houseDevice.Id.Ip = incoming.Ip
	}

	timeAway := s.deviceDetectState(houseDevice.LastSeen)
	if timeAway > *timeAwaySeconds {
		log.Println(fmt.Sprintf("Device: %s has returned after %d seconds", houseDevice.Name, timeAway))
		if houseDevice.Person {
			if *debug {
				log.Printf("Notification: %s, %s", houseDevice.Name, fmt.Sprintf("has returned to %s.", houseDevice.Home))
			} else {
				err := notifications.SendNotification(houseDevice.Name, fmt.Sprintf("has returned to %s.", houseDevice.Home), houseDevice.Home)
				if err != nil {
					return err
				}
			}
		}
	}

	if houseDevice.Person {
		homeKey := fmt.Sprintf("%s%s", homePrefix, houseDevice.Home)
		houseStatus, err := s.etcdClient.Get(context.Background(), homeKey)
		if err != nil {
			log.Panic(err.Error())
		}

		if houseStatus.Count == 0 {
			homeKey := fmt.Sprintf("%s%s", homePrefix, houseDevice.Home)
			_, err = s.etcdClient.Put(context.Background(), homeKey, "false")
			if err != nil {
				log.Panic(err.Error())
			}
		} else if val, err := strconv.ParseBool(string(houseStatus.Kvs[0].Value)); val && err == nil {
			homeKey := fmt.Sprintf("%s%s", homePrefix, houseDevice.Home)
			_, err = s.etcdClient.Put(context.Background(), homeKey, "false")
			if err != nil {
				log.Panic(err.Error())
			}
			if *debug {
				log.Println("House no longer empty")
			} else {
				err := notifications.SendNotification(houseDevice.Home, "No longer Empty", houseDevice.Home)
				if err != nil {
					return err
				}
			}
		}
	}

	houseDevice.Away = false
	houseDevice.LastSeen = int64(time.Now().Unix())
	if incoming.Mac != "" && incoming.Mac == houseDevice.Id.Mac {
		err := s.writeNetworkDevice(houseDevice)
		if err != nil {
			log.Printf("Error saving to ETCD: %s", err.Error())
		}
	}
	return nil
}

// searchForOverlappingDevices Checks if reported device
func (s *Server) searchForOverlappingDevices(in *pb.AddressRequest) (*bool, error) {
	devices, err := s.readNetworkConfig()
	if err != nil {
		return nil, err
	}
	in.Mac = in.Ip

	found := false
	for _, v := range devices {
		// on same home and same ip, report it as the one with a mac
		if v.Id.Ip == in.Ip && v.Home == in.Home && strings.Contains(v.Id.Mac, ":") {
			in.Mac = v.Id.Mac
			found = true
			return &found, err
		}
	}

	//if found {
	//	etcdKey := strings.ReplaceAll(strings.ReplaceAll(in.Ip, ".", "_"), ":", "_")
	//	_, err = s.etcdClient.Delete(context.Background(), fmt.Sprintf("%s%s", devicesPrefix, etcdKey))
	//	if err != nil {
	//		log.Println(err.Error())
	//	}
	//	_, err = s.etcdClient.Delete(context.Background(), fmt.Sprintf("%s%s", devicesPrefix, in.Ip))
	//	if err != nil {
	//		log.Println(err.Error())
	//	}
	//}

	return &found, err
}

// Address Handler for receiving IP/MAC requests
func (s *Server) Address(ctx context.Context, in *pb.AddressRequest) (*pb.Reply, error) {
	incoming := in
	clientIpFullIp, _ := peer.FromContext(ctx)
	clientFullIpString := clientIpFullIp.Addr.String()
	clientIpV4 := clientFullIpString[:strings.Index(clientFullIpString, ":")]
	// assuming mac is empty as its the clients own ip
	if clientIpV4 == incoming.Ip && incoming.Mac == "" {
		return &pb.Reply{Acknowledged: true}, nil
	}
	if incoming.Mac == "" && incoming.Home != "" {
		incoming.Mac = fmt.Sprintf("%s/%s", incoming.Home, strings.ReplaceAll(in.Ip, ".", "_"))
	}
	opts := []etcdv3.OpOption{
		etcdv3.WithLimit(1),
	}
	item, err := s.etcdClient.Get(context.Background(), fmt.Sprintf("%s%s", devicesPrefix, in.Mac), opts...)
	if err != nil {
		return nil, err
	}
	if item.Count == 0 {
		err := s.newDevice(in)
		if err != nil {
			return nil, err
		}
	} else {
		strDevice := item.Kvs[0].Value
		var exDevice *device
		err = yaml.Unmarshal(strDevice, &exDevice)
		if err != nil {
			return nil, err
		}

		err = s.existingDevice(exDevice, incoming)
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
					err = notifications.SendNotification(device.Name, command.Command, "devices")
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

func (s *Server) isHouseEmpty(home string) bool {
	houseEmpty := true
	devices, err := s.readNetworkConfig()
	if err != nil {
		log.Println("Failed to read from etcd")
	}
	for _, device := range devices {
		if !device.Away && device.Person && device.Home == home {
			houseEmpty = false
		}
	}
	return houseEmpty
}

func (s *Server) iotStatusManager() error {
	gHouseEmpty, err := s.readHomesConfig()
	if err != nil {
		return err
	}
	for home, empty := range gHouseEmpty {
		if houseEmpty := s.isHouseEmpty(home); houseEmpty != *empty {
			_, err := s.etcdClient.Put(context.Background(), fmt.Sprintf("%s%s", homePrefix, home), strconv.FormatBool(houseEmpty))
			if err != nil {
				log.Println(err)
				return err
			}
			if *debug {
				log.Printf("House (%s) is Empty(%v)", home, houseEmpty)
			} else {
				err := notifications.SendNotification("House Empty", fmt.Sprintf("No Humans in %s", home), home)
				if err != nil {
					log.Println(err)
					return err
				}
				if strings.Contains(home, "wst") {
					_, err = assistant.Call("Everyone's away")
					if err != nil {
						log.Println(err)
						return err
					}
				}
			}
		}
	}

	return nil
}

func (s *Server) deviceManager() error {
	devices, err := s.readNetworkConfig()
	if err != nil {
		log.Println("Failed to read from etcd")
	}
	for _, device := range devices {
		timeAway := s.deviceDetectState(device.LastSeen)
		if timeAway > *timeAwaySeconds && !device.Away {
			log.Println(fmt.Sprintf("Device: %s has left after %d seconds", device.Name, timeAway))
			device.Away = true
			if device.Person {
				if !*debug {
					err := notifications.SendNotification(device.Name, "Has left the house", device.Home)
					if err != nil {
						return nil
					}
				} else {
					log.Printf("Notification: %s: %s", device.Name, "Has left the house")
				}
			}
			err = s.writeNetworkDevice(device)
			if err != nil {
				return nil
			}
			err := s.iotStatusManager()
			if err != nil {
				return nil
			}
		}
	}
	return nil
}
