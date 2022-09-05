package house

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/beaujr/nmap_prometheus/etcd"
	pb "github.com/beaujr/nmap_prometheus/proto"
	"github.com/ghodss/yaml"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/robfig/cron/v3"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//IOT iotDevices
var (
	timeAwaySeconds    = flag.Int64("timeout", 300, "")
	bleTimeAwaySeconds = flag.Int64("bleTimeout", 15, "")
	networkConfigFile  = flag.String("config", "config/devices.yaml", "Path to config file")
	bleConfigFile      = flag.String("bleconfig", "config/ble_devices.yaml", "Path to config file")
	etcdServers        = flag.String("etcdServers", "192.168.1.216:2379", "Comma Separated list of etcd servers")
	debug              = flag.Bool("debug", false, "Debug mode")
	cqEnabled          = flag.Bool("cq", false, "Command Queue Enabled")
	newDeviceIsPerson  = flag.Bool("newDeviceIsPerson", false, "Track new devices as people")
)

var bleDevices = []*pb.BleDevices{}

var syncStatusWithGA = time.Hour.Seconds()
var metrics map[string]prometheus.Gauge
var devicesPrefix = "/devices/"
var homePrefix = "/homes/"
var BlesPrefix = "/bles/"
var tcPrefix = "/cq/"
var peoplePrefix = "/people/"
var notificationsPrefix = "/notifications/"

var (
	peopleHome = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "home_detector_people_home",
		Help: "The total number of houseDevices at home",
	})
)

// HomeManager manages devices and metric collection
type HomeManager interface {
	deviceDetectState(phone int64) int64
	deviceManager() error
	isDeviceOn(iot *pb.Devices) (bool, error)
	IsHouseEmpty(home string) bool
	httpHealthCheck(url string) bool
	iotStatusManager() error
	RecordMetrics()
	Devices(w http.ResponseWriter, req *http.Request)
	People(w http.ResponseWriter, req *http.Request)
	HomeEmptyState(w http.ResponseWriter, req *http.Request)
}

// Server is an implementation of the proto HomeDetectorServer
type Server struct {
	pb.UnimplementedHomeDetectorServer
	EtcdClient         etcdv3.KV
	AssistantClient    GoogleAssistant
	NotificationClient Notifier
}

func writeConfig(data []byte, filename string) error {
	err := ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

// TimeCommands implements sort.Interface by providing Less and using the Len and
type TimeCommands []*pb.TimedCommands

// ByExecutedAt implements TimeCommands
type ByExecutedAt struct{ TimeCommands }

// Len returns length of TimeCommands Array
func (s TimeCommands) Len() int { return len(s) }

// Swap sorts the array using equivalent comparison
func (s TimeCommands) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// Less sorts the array using less than comparison
func (s ByExecutedAt) Less(i, j int) bool {
	return s.TimeCommands[i].Executeat < s.TimeCommands[j].Executeat
}
func NewCustomServer(e etcdv3.KV, g GoogleAssistant, n Notifier) Server {
	s := Server{
		UnimplementedHomeDetectorServer: pb.UnimplementedHomeDetectorServer{},
		EtcdClient:                      e,
		AssistantClient:                 g,
		NotificationClient:              n,
	}
	createCrons(&s)
	return s
}

func createCrons(server *Server) {
	c := cron.New(cron.WithSeconds())
	c.AddFunc("0 * * * * *", func() {
		err := server.deviceManager()
		if err != nil {
			log.Println(err)
		}
		err = server.iotStatusManager()
		if err != nil {
			log.Println(err)
		}
	})
	c.AddFunc("* */1 * * * *", func() {
		knowDevices, err := server.ReadNetworkConfig()
		if err != nil {
			log.Printf(err.Error())
		}
		for _, val := range knowDevices {
			server.RegisterMetric(*val)
		}
	})
	if *cqEnabled {
		c.AddFunc("*/10 * * * * *", func() {
			err := server.processTimedCommandQueue()
			if err != nil {
				log.Println(err)
			}
		})
	}
	c.Start()
	return
}

// NewServer new instance of HomeManager
func NewServer() HomeManager {
	etcdClient := etcd.NewClient(strings.Split(*etcdServers, ","))
	assistantClient := NewAssistant()
	notifyClient := NewNotifier(etcdClient)

	server := &Server{EtcdClient: etcdClient, AssistantClient: assistantClient, NotificationClient: notifyClient}
	_, err := server.ReadNetworkConfig()
	if err != nil {
		log.Println(err)
	}
	// importConfig to etcd
	bleDevices, _ := readBleConfig(*bleConfigFile)
	for _, item := range bleDevices {
		_ = server.writeBleDevice(item)
	}

	// Bluetooth
	_, err = server.ReadBleConfig()
	if err != nil {
		log.Println(err)
	}

	knowDevices, err := server.ReadNetworkConfig()
	if err != nil {
		log.Printf(err.Error())
	}
	homes := make([]string, 0)
	for _, item := range knowDevices {
		log.Println(fmt.Sprintf("Registering metric for %s", item.Name))
		server.RegisterMetric(*item)
		homes = append(homes, item.Home)
	}

	for _, home := range homes {
		homeKey := fmt.Sprintf("%s%s", homePrefix, home)
		_, err := server.EtcdClient.Put(context.Background(), homeKey, strconv.FormatBool(server.IsHouseEmpty(home)))
		if err != nil {
			log.Panic(err.Error())
		}
	}
	createCrons(server)
	server.RecordMetrics()
	return server
}

// Devices API endpoint to determine devices status
func (s *Server) Devices(w http.ResponseWriter, req *http.Request) {
	devices := make([]*pb.Devices, 0)
	items, err := s.ReadNetworkConfig()
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
	people := make([]*pb.Devices, 0)
	devices, err := s.ReadNetworkConfig()
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
	homes, err := s.ReadHomesConfig()
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
func init() {
	metrics = make(map[string]prometheus.Gauge)
}

func (s *Server) RegisterMetric(item pb.Devices) {
	s.AddMetric("", &item)
	s.AddMetric("lastseen", &item)
}

func (s *Server) AddMetric(key string, item *pb.Devices) prometheus.Gauge {
	metricsKey := fmt.Sprintf("%s_%s", item.Name, key)
	if key == "" {
		metricsKey = item.Name
	}
	if metrics[metricsKey] == nil {
		metricNamespace := fmt.Sprintf("home_detector_device_%s", key)
		if key == "" {
			metricNamespace = "home_detector_device"
		}
		metrics[metricsKey] = promauto.NewGauge(prometheus.GaugeOpts{
			Name: metricNamespace,
			Help: fmt.Sprintf("Device in home %s", key),
			ConstLabels: prometheus.Labels{
				"name":   strings.ReplaceAll(item.Name, " ", "_"),
				"mac":    item.Id.Mac,
				"ip":     item.Id.Ip,
				"home":   item.Home,
				"person": strconv.FormatBool(item.Person),
			},
		})
	}
	return metrics[metricsKey]
}

func (s *Server) RecordMetrics() {
	go func() {
		for {
			knowDevices, err := s.ReadNetworkConfig()
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
				s.AddMetric("", item).Set(float64(state))
				s.AddMetric("lastseen", item).Set(float64(item.GetLastSeen()))
				s.AddMetric("distance", item).Set(float64(item.GetLatency()))
			}
			time.Sleep(2 * time.Second)
		}
	}()
}

func (s *Server) callAssistant(command string) (*string, error) {
	return s.AssistantClient.Call(command)
}

func (s *Server) deviceDetectState(deviceLastSeen int64) int64 {
	now := int64(time.Now().Unix())
	lastSeen := now - deviceLastSeen
	return lastSeen
}

func (s *Server) newBleDevice(in *pb.StringRequest) error {
	newDevice := pb.BleDevices{
		Id:       in.Key,
		LastSeen: int64(time.Now().Unix()),
		Commands: make([]*pb.Commands, 0),
	}

	log.Println(fmt.Printf("New BLE Device: %s", in.Key))

	bleDevices = append(bleDevices, &newDevice)
	_, err := uniqueBle(bleDevices)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) newDevice(in *pb.AddressRequest, home string) error {
	name := in.Ip
	if in.Mac != "" {
		name = in.Mac
	}
	vendor := "unknown"
	if in.Mac != in.Ip && strings.Contains(in.Mac, ":") {
		macVendor, err := GetManufacturer(in.Mac)
		if macVendor != nil {
			vendor = *macVendor
		}
		if err != nil {
			log.Printf(err.Error())
			vendor = name
		}

	}

	newDevice := pb.Devices{
		Name:         strings.ReplaceAll(strings.ReplaceAll(name, ".", "_"), ":", "_"),
		Id:           &pb.NetworkId{Ip: in.Ip, Mac: in.Mac, UUID: name},
		Away:         false,
		LastSeen:     int64(time.Now().Unix()),
		Person:       *newDeviceIsPerson,
		Command:      "",
		Manufacturer: vendor,
		Home:         home,
	}

	log.Println(fmt.Printf("New Device: %s", name))

	err := s.writeNetworkDevice(&newDevice)
	if err != nil {
		log.Printf("Error saving to ETCD: %s", err.Error())
	}

	err = s.NotificationClient.SendNotification(fmt.Sprintf("New Device in %s (%s)", newDevice.Home, newDevice.Id.Ip), newDevice.Manufacturer, newDevice.Home)
	if err != nil {
		log.Printf("Error sending notification: %s", err.Error())
	}
	s.RegisterMetric(newDevice)
	return nil
}

func (s *Server) existingDevice(houseDevice *pb.Devices, incoming *pb.AddressRequest, home string) error {
	if incoming.Mac != "" {
		houseDevice.Id.Mac = incoming.Mac
	}

	if incoming.Mac == "" && incoming.Ip == houseDevice.Id.UUID {
		houseDevice.Id.Ip = incoming.Ip
	}

	if incoming.Ip != houseDevice.Id.Ip {
		houseDevice.Id.Ip = incoming.Ip
	}

	if home != houseDevice.Home {
		houseDevice.Home = home
		message := fmt.Sprintf("%s has moved to %s", houseDevice.Name, houseDevice.Home)
		err := s.NotificationClient.SendNotification(houseDevice.Home, message, houseDevice.Home)
		if err != nil {
			return err
		}
	}

	timeAway := s.deviceDetectState(houseDevice.LastSeen)
	if timeAway > *timeAwaySeconds {
		log.Println(fmt.Sprintf("Device: %s has returned after %d seconds", houseDevice.Name, timeAway))
		topic := fmt.Sprintf("%s_devices", houseDevice.Home)
		title := houseDevice.GetId().GetMac()
		if houseDevice.Person {
			topic = houseDevice.GetHome()
			title = houseDevice.GetName()
		}
		if houseDevice.Person {
			topic = houseDevice.GetHome()
		}
		err := s.NotificationClient.SendNotification(title, fmt.Sprintf("has returned to %s.", houseDevice.Home), topic)
		if err != nil {
			return err
		}
	}

	if houseDevice.Person {
		err := s.processPerson(houseDevice)
		if err != nil {
			return nil
		}
	}
	houseDevice.Away = false
	houseDevice.LastSeen = int64(time.Now().Unix())

	houseDevice.Latency = incoming.GetDistance()
	if incoming.Mac != "" && incoming.Mac == houseDevice.Id.Mac {
		err := s.writeNetworkDevice(houseDevice)
		if err != nil {
			log.Printf("Error saving to ETCD: %s", err.Error())
		}
	}
	return nil
}

// searchForOverlappingDevices Checks if reported device
func (s *Server) searchForOverlappingDevices(in *pb.AddressRequest, home string) (*bool, error) {
	devices, err := s.ReadNetworkConfig()
	if err != nil {
		return nil, err
	}
	in.Mac = in.Ip
	found := false
	for _, v := range devices {
		// on same home and same ip, report it as the one with a mac
		if v.Id.Ip == in.Ip && v.Home == home && strings.Contains(v.Id.Mac, ":") {
			in.Mac = v.Id.Mac
			found = true
			return &found, err
		}
	}
	return &found, err
}

func (s *Server) ProcessIncomingAddress(ctx context.Context, in *pb.AddressRequest) (*pb.Reply, error) {
	incoming := in
	headers, _ := metadata.FromIncomingContext(ctx)
	home := "unknown"
	val := headers.Get("home")
	if len(val) > 0 {
		home = val[0]
	}
	if incoming.Mac == "" && home != "" {
		incoming.Mac = fmt.Sprintf("%s/%s", home, strings.ReplaceAll(in.Ip, ".", "_"))
	}
	opts := []etcdv3.OpOption{
		etcdv3.WithLimit(1),
		etcdv3.WithKeysOnly(),
	}
	item, err := s.EtcdClient.Get(ctx, fmt.Sprintf("%s%s", devicesPrefix, in.Mac), opts...)
	if err != nil {
		return nil, err
	}
	if item.Count == 0 {
		err := s.newDevice(in, home)
		if err != nil {
			return nil, err
		}
	} else {
		strDevice := item.Kvs[0].Value
		var exDevice *pb.Devices
		err = yaml.Unmarshal(strDevice, &exDevice)
		if err != nil {
			return nil, err
		}

		err = s.existingDevice(exDevice, incoming, home)
		if err != nil {
			return nil, err
		}
	}

	return &pb.Reply{Acknowledged: true}, nil
}
func (s *Server) grpcHitsMetrics(promMetric string, name string, itemCount int) {
	if metrics[promMetric] == nil {
		metrics[promMetric] = promauto.NewGauge(prometheus.GaugeOpts{
			Name: "home_detector_grpc_endpoint_items",
			Help: "Number of calls to server endpoint",
			ConstLabels: prometheus.Labels{
				"name": name,
			},
		})
		metrics[promMetric].Set(0)
	}
	metrics[promMetric].Add(float64(itemCount))
}

func (s *Server) grpcPrometheusMetrics(ctx context.Context, promMetric string, name string) {
	if metrics[promMetric] == nil {
		metrics[promMetric] = promauto.NewGauge(prometheus.GaugeOpts{
			Name: "home_detector_grpc_endpoint",
			Help: "Number of calls to server endpoint",
			ConstLabels: prometheus.Labels{
				"name": name,
			},
		})
		metrics[promMetric].Set(0)
	}
	metrics[promMetric].Add(1)
	headers, _ := metadata.FromIncomingContext(ctx)
	home := "unknown"
	val := headers.Get("home")
	if len(val) > 0 {
		home = val[0]
	}
	if val := headers.Get("client"); len(val) > 0 {
		agentType := "nmap"
		if strings.Compare("Ack", name) == 0 {
			agentType = "ble"
		}
		promClientMetric := fmt.Sprintf("%s_client", val[0])
		if metrics[promClientMetric] == nil {
			metrics[promClientMetric] = promauto.NewGauge(prometheus.GaugeOpts{
				Name: "home_detector_grpc_clients",
				Help: "Number of calls to server endpoint",
				ConstLabels: prometheus.Labels{
					"name": val[0],
					"home": home,
					"type": agentType,
				},
			})
		}
		metrics[promClientMetric].Set(float64(time.Now().Unix()))
	}
}

func (s *Server) getBLEById(id *string) (*pb.BleDevices, error) {
	opts := []etcdv3.OpOption{
		etcdv3.WithLimit(1),
	}
	log.Println(fmt.Sprintf("%s%s", BlesPrefix, *id))
	item, err := s.EtcdClient.Get(context.Background(), fmt.Sprintf("%s%s", BlesPrefix, *id), opts...)
	if err != nil {
		return nil, err
	}
	found := item.Count == 1
	if found {
		strDevice := item.Kvs[0].Value
		var device *pb.BleDevices
		err = yaml.Unmarshal(strDevice, &device)
		if err != nil {
			return nil, err
		}
		return device, nil
	}
	return nil, nil
}

func (s *Server) processIncomingBleAddress(ctx context.Context, in *pb.BleRequest) (*bool, error) {
	device, err := s.getBLEById(&in.Key)
	if err != nil {
		return nil, err
	}
	found := device != nil
	if !found {
		return &found, nil
	}
	if found {
		lastSeen := s.deviceDetectState(device.LastSeen)
		device.LastSeen = time.Now().Unix()
		device.Distance = in.Distance
		err := s.writeBleDevice(device)
		if err != nil {
			log.Printf("Error updating BLE device: %s", device.Id)
		}
		if lastSeen > *bleTimeAwaySeconds {
			log.Printf("BLE: %s (%s) detected", device.Name, device.Id)
			for _, command := range device.Commands {
				err = s.createTimedCommand(command.Timeout, device.Id, command.Id, command.Command, device.Id)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return &found, nil
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

func (s *Server) isDeviceOn(iot *pb.Devices) (bool, error) {
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

func (s *Server) IsHouseEmpty(home string) bool {
	houseEmpty := true
	devices, err := s.ReadNetworkConfig()
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

func (s *Server) deviceManager() error {
	devices, err := s.ReadNetworkConfig()
	if err != nil {
		log.Println("Failed to read from etcd")
	}
	for _, device := range devices {
		timeAway := s.deviceDetectState(device.LastSeen)
		if timeAway > *timeAwaySeconds && !device.Away {
			log.Println(fmt.Sprintf("Device: %s has left after %d seconds", device.Name, timeAway))
			device.Away = true
			topic := fmt.Sprintf("%s_devices", device.Home)
			title := device.GetId().GetMac()
			if device.Person {
				topic = device.GetHome()
				title = device.GetName()
			}
			err := s.NotificationClient.SendNotification(title, "Has left the house", topic)
			if err != nil {
				return nil
			}
			err = s.writeNetworkDevice(device)
			if err != nil {
				return nil
			}
		}
	}
	return nil
}
