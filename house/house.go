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
	dto "github.com/prometheus/client_model/go"
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
	etcdServers        = flag.String("etcdServers", "192.168.1.232:2379", "Comma Separated list of etcd servers")
	debug              = flag.Bool("debug", false, "Debug mode")
	cqEnabled          = flag.Bool("cq", false, "Command Queue Enabled")
	newDeviceIsPerson  = flag.Bool("newDeviceIsPerson", false, "Track new devices as people")
)

var bleDevices = []*pb.BleDevices{}

var syncStatusWithGA = time.Hour.Seconds()
var metrics map[string]*prometheus.GaugeVec
var devicesPrefix = "/devices/"
var homePrefix = "/homes/"
var AlivePrefix = "/alive/"
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
	loadMetrics()
	Devices(w http.ResponseWriter, req *http.Request)
	People(w http.ResponseWriter, req *http.Request)
	HomeEmptyState(w http.ResponseWriter, req *http.Request)
}

// Server is an implementation of the proto HomeDetectorServer
type Server struct {
	pb.UnimplementedHomeDetectorServer
	Kv                 etcdv3.KV
	AssistantClient    GoogleAssistant
	EtcdClient         *etcdv3.Client
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
		Kv:                              e,
		AssistantClient:                 g,
		NotificationClient:              n,
	}
	createCrons(&s)
	s.loadMetrics()
	return s
}

func createCrons(server *Server) {
	c := cron.New(cron.WithSeconds())
	//c.AddFunc("*/2 * * * * *", func() {
	//	err := server.deviceManager()
	//	if err != nil {
	//		log.Println(err)
	//	}
	//	err = server.iotStatusManager()
	//	if err != nil {
	//		log.Println(err)
	//	}
	//})
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
	client, etcdClient := etcd.NewClient(strings.Split(*etcdServers, ","))
	assistantClient := NewAssistant()
	notifyClient := NewNotifier(etcdClient)

	server := &Server{Kv: etcdClient, AssistantClient: assistantClient, NotificationClient: notifyClient, EtcdClient: client}
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
	bles, err := server.ReadBleConfig()
	if err != nil {
		log.Println(err)
	}
	for _, item := range bles {
		server.RegisterBleMetric(item, "etcd")
	}

	knowDevices, err := server.ReadNetworkConfig()
	if err != nil {
		log.Printf(err.Error())
	}
	homes := make([]string, 0)
	for _, item := range knowDevices {
		homes = append(homes, item.Home)
	}

	for _, home := range homes {
		homeKey := fmt.Sprintf("%s%s", homePrefix, home)
		_, err := server.Kv.Put(context.Background(), homeKey, strconv.FormatBool(server.IsHouseEmpty(home)))
		if err != nil {
			log.Panic(err.Error())
		}
	}
	createCrons(server)
	server.loadMetrics()
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
	metrics = make(map[string]*prometheus.GaugeVec)
	hdd := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "home_detector_device",
			Help: "Device in home",
		},
		[]string{
			"name", "mac", "home", "person",
		},
	)
	lastseen := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "home_detector_device_lastseen",
			Help: "lastseen device in home",
		},
		[]string{
			"name", "mac", "home", "person",
		},
	)
	distance := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "home_detector_device_distance",
			Help: "distance device in home",
		},
		[]string{
			"name", "mac", "home", "person",
		},
	)
	bledistance := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "home_detector_ble_distance",
			Help: "distance device in home",
		},
		[]string{
			"name", "mac", "home", "agent",
		},
	)
	grpc := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "grpc_address_count",
			Help: "Amount of times GRPC Endpoint hit",
		},
		[]string{
			"name",
		},
	)
	grpcEndpoint := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "home_detector_grpc_endpoint",
			Help: "agents hitting endpoint",
		},
		[]string{
			"name",
		},
	)
	grpcAgentEndpoint := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "home_detector_grpc_clients",
		},
		[]string{
			"name", "home", "type",
		},
	)
	cq := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "home_detector_ble_device",
		},
		[]string{
			"name", "command",
		},
	)
	metrics["lastseen"] = lastseen
	metrics["hdd"] = hdd
	metrics["distance"] = distance
	metrics["grpc"] = grpc
	metrics["grpcEndpoint"] = grpcEndpoint
	metrics["grpcAgentEndpoint"] = grpcAgentEndpoint
	metrics["cq"] = cq
	metrics["bledistance"] = bledistance
	for _, v := range metrics {
		prometheus.MustRegister(v)
	}

}

func (s *Server) RegisterMetric(item *pb.Devices) {
	away := 0
	if item.GetAway() {
		away = 1
	}
	metrics["hdd"].WithLabelValues(item.Name, item.GetId().GetMac(), item.GetHome(), strconv.FormatBool(item.GetPerson())).Set(float64(away))
	metrics["lastseen"].WithLabelValues(item.Name, item.GetId().GetMac(), item.GetHome(), strconv.FormatBool(item.GetPerson())).Set(float64(item.GetLastSeen()))
	metrics["distance"].WithLabelValues(item.Name, item.GetId().GetMac(), item.GetHome(), strconv.FormatBool(item.GetPerson())).Set(float64(item.GetLatency()))
}

func (s *Server) RegisterBleMetric(item *pb.BleDevices, agent string) {
	metrics["bledistance"].WithLabelValues(item.Name, item.GetId(), item.GetHome(), agent).Set(float64(item.GetDistance()))
	metrics["lastseen"].WithLabelValues(item.Name, item.GetId(), item.GetHome(), "false").Set(float64(item.GetLastSeen()))
}

func (s *Server) loadMetrics() {
	knowDevices, err := s.ReadNetworkConfig()
	if err != nil {
		log.Printf(err.Error())
	}
	for _, item := range knowDevices {
		state := 1
		if item.Away {
			state = 0
		}
		metrics["hdd"].WithLabelValues(item.Name, item.GetId().GetMac(), item.GetHome(), strconv.FormatBool(item.GetPerson())).Set(float64(state))
		s.RecordLastSeenMetric(item, item.GetLastSeen())
		s.RecordDistanceMetric(item, item.GetLatency())
	}
	// Bluetooth
	bles, err := s.ReadBleConfig()
	if err != nil {
		log.Println(err)
	}
	for _, item := range bles {
		s.RegisterBleMetric(item, "etcd")
	}
}

func (s *Server) RecordLastSeenMetric(item *pb.Devices, timestamp int64) {
	metrics["lastseen"].WithLabelValues(item.Name, item.GetId().GetMac(), item.GetHome(), strconv.FormatBool(item.GetPerson())).Set(float64(timestamp))
	timeAway := s.deviceDetectState(item.LastSeen)
	if timeAway > *timeAwaySeconds {
		metrics["hdd"].WithLabelValues(item.Name, item.GetId().GetMac(), item.GetHome(), strconv.FormatBool(item.GetPerson())).Set(float64(0))
	} else {
		metrics["hdd"].WithLabelValues(item.Name, item.GetId().GetMac(), item.GetHome(), strconv.FormatBool(item.GetPerson())).Set(float64(1))
	}
}

func (s *Server) RecordDistanceMetric(item *pb.Devices, distance float32) {
	metrics["distance"].WithLabelValues(item.Name, item.GetId().GetMac(), item.GetHome(), strconv.FormatBool(item.GetPerson())).Set(float64(distance))
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

	err := s.WriteNetworkDevice(&newDevice)
	if err != nil {
		log.Printf("Error saving to ETCD: %s", err.Error())
	}
	err = s.NotificationClient.SendNotification(fmt.Sprintf("New Device in %s (%s)", newDevice.Home, newDevice.Id.Ip), newDevice.Manufacturer, newDevice.Home)
	if err != nil {
		log.Printf("Error sending notification: %s", err.Error())
	}
	s.RegisterMetric(&newDevice)
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

	if houseDevice.GetPerson() {
		err := s.processPerson(houseDevice)
		if err != nil {
			return nil
		}
	}
	houseDevice.LastSeen = int64(time.Now().Unix())
	houseDevice.Latency = incoming.GetDistance()
	if incoming.Mac != "" && incoming.Mac == houseDevice.Id.Mac {
		err := s.WriteNetworkDevice(houseDevice)
		if err != nil {
			log.Printf("Error saving to ETCD: %s", err.Error())
		}
	}

	return nil
}

func (s *Server) GrantLease(mac, home string, ttl int64) error {
	ctx := context.Background()
	leaseExists := false
	leases, err := s.EtcdClient.Lease.Leases(ctx)
	if err != nil {
		return err
	}
	for _, lease := range leases.Leases {
		leaseTTL, err := s.EtcdClient.Lease.TimeToLive(ctx, lease.ID, etcdv3.WithAttachedKeys())
		if err != nil {
			return err
		}
		for _, key := range leaseTTL.Keys {
			if strings.Compare(fmt.Sprintf("%s%s", AlivePrefix, mac), string(key)) == 0 {
				leaseExists = true
				_, err := s.EtcdClient.Lease.KeepAliveOnce(context.Background(), leaseTTL.ID)
				if err != nil {
					return err
				}

			}
		}
		if leaseExists {
			break
		}
	}
	if !leaseExists {
		lease, err := s.EtcdClient.Lease.Grant(ctx, ttl)
		if err != nil {
			return err
		}
		key := fmt.Sprintf("%s%s", AlivePrefix, mac)
		_, err = s.EtcdClient.Put(ctx, key, home, etcdv3.WithLease(lease.ID))
		if err != nil {
			return err
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
	}

	err := s.GrantLease(in.Mac, home, *timeAwaySeconds)
	if err != nil {
		return nil, err
	}

	item, err := s.Kv.Get(ctx, fmt.Sprintf("%s%s", devicesPrefix, in.Mac), opts...)
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
		s.RecordDistanceMetric(exDevice, incoming.GetDistance())
		s.RecordLastSeenMetric(exDevice, time.Now().Unix())
	}

	return &pb.Reply{Acknowledged: true}, nil
}
func (s *Server) GrpcHitsMetrics(promMetric string, name string, itemCount int) {
	metrics["grpc"].WithLabelValues(name).Add(float64(itemCount))
}

func (s *Server) GrpcPrometheusMetrics(ctx context.Context, promMetric string, name string) {
	metrics["grpcEndpoint"].WithLabelValues(name).Add(1)
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
		metrics["grpcAgentEndpoint"].WithLabelValues(val[0], home, agentType).Set(float64(time.Now().Unix()))
	}
}

func (s *Server) getBLEById(id *string) (*pb.BleDevices, error) {
	opts := []etcdv3.OpOption{
		etcdv3.WithLimit(1),
	}
	log.Println(fmt.Sprintf("%s%s", BlesPrefix, *id))
	item, err := s.Kv.Get(context.Background(), fmt.Sprintf("%s%s", BlesPrefix, *id), opts...)
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

	device.LastSeen = time.Now().Unix()
	device.Distance = in.Distance
	err = s.writeBleDevice(device)
	if err != nil {
		log.Printf("Error updating BLE device: %s", device.Id)
	}
	headers, _ := metadata.FromIncomingContext(ctx)
	if val := headers.Get("client"); len(val) > 0 {
		s.RegisterBleMetric(device, val[0])
	}

	if len(device.GetCommands()) > 0 {
		tcs, err := s.getTcByOwner(in.GetKey())
		if err != nil {
			return nil, err
		}
		if len(tcs) > 0 {
			err = s.NotificationClient.SendNotification("Device already detected", fmt.Sprintf("Device Left on %s.", device.GetName()), device.GetHome())
			if err != nil {
				return nil, err
			}
			return &found, nil
		}
		for _, command := range device.Commands {
			err = s.createTimedCommand(command.Timeout, device.Id, command.Id, command.Command, device.Id)
			if err != nil {
				return nil, err
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
	devices, err := s.ReadNetworkConfig()
	if err != nil {
		log.Println("Failed to read from etcd")
	}
	for _, device := range devices {
		if !device.Away && device.Person && device.Home == home {
			return false
		}
	}
	return true
}

func (s *Server) deviceManager() error {
	lastSeenItems := GetMetricValue(metrics["lastseen"])
	deviceStateItems := GetMetricValue(metrics["hdd"])
	for _, state := range deviceStateItems {
		if int(*state.Gauge.Value) == 1 {
			for _, item := range lastSeenItems {
				stateId := ""
				for _, lbl := range state.GetLabel() {
					if lbl.GetName() == "mac" {
						stateId = lbl.GetValue()
					}
				}
				if len(stateId) == 0 {
					continue
				}
				timeAway := s.deviceDetectState(int64(*item.Gauge.Value))
				if timeAway > *timeAwaySeconds {
					id := ""
					for _, lbl := range item.GetLabel() {
						if lbl.GetName() == "mac" {
							id = lbl.GetValue()
						}
					}
					if len(stateId) == 0 {
						continue
					}
					if stateId == id {
						device, err := s.GetDevice(id)
						if err != nil {
							return err
						}
						if device.GetPerson() {
							continue
						}
						log.Println(fmt.Sprintf("Device: %s has left after %d seconds", device.Name, timeAway))
						device.Away = true
						topic := fmt.Sprintf("%s_devices", device.Home)
						title := device.GetId().GetMac()
						if device.Person {
							topic = device.GetHome()
							title = device.GetName()
						}
						err = s.NotificationClient.SendNotification(title, "Has left the house", topic)
						if err != nil {
							return nil
						}
						err = s.WriteNetworkDevice(device)
						if err != nil {
							return nil
						}
						metrics["hdd"].WithLabelValues(device.Name, device.GetId().GetMac(), device.GetHome(), strconv.FormatBool(device.GetPerson())).Set(0)

					}
				}
			}
		}

	}
	return nil
}
func GetBleDeviceMetricByMac(mac string) *dto.Metric {
	return findMetricByLabelAndValue("mac", mac, metrics["bledistance"])
}

func findMetricByLabelAndValue(name, value string, metric *prometheus.GaugeVec) *dto.Metric {
	cMetrics := []*dto.Metric{}
	collect(metric, func(m dto.Metric) {
		for _, lbl := range m.GetLabel() {
			if lbl.GetValue() == value && lbl.GetName() == name {
				cMetrics = append(cMetrics, &m)
			}
		}
	})
	if len(cMetrics) == 0 {
		return nil
	}
	return cMetrics[0]
}

func GetDeviceMetricByMac(mac string) *dto.Metric {
	return findMetricByLabelAndValue("mac", mac, metrics["lastseen"])
}

func GetMetricValue(col prometheus.Collector) []*dto.Metric {
	cMetrics := []*dto.Metric{}
	collect(col, func(m dto.Metric) {
		cMetrics = append(cMetrics, &m)
	})
	return cMetrics
}

// collect calls the function for each metric associated with the Collector
func collect(col prometheus.Collector, do func(dto.Metric)) {
	c := make(chan prometheus.Metric)
	go func(c chan prometheus.Metric) {
		col.Collect(c)
		close(c)
	}(c)
	for x := range c { // eg range across distinct label vector values
		m := dto.Metric{}
		_ = x.Write(&m)
		do(m)
	}
}
