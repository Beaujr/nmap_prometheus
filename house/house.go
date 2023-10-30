package house

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/beaujr/nmap_prometheus/agent"
	"github.com/beaujr/nmap_prometheus/etcd"
	pb "github.com/beaujr/nmap_prometheus/proto"
	"github.com/ghodss/yaml"
	"github.com/robfig/cron/v3"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// IOT iotDevices
var (
	TimeAwaySeconds    = flag.Int64("timeout", 300, "")
	BleTimeAwaySeconds = flag.Int64("bleTimeout", 15, "")
	networkConfigFile  = flag.String("config", "config/devices.yaml", "Path to config file")
	bleConfigFile      = flag.String("bleconfig", "config/ble_devices.yaml", "Path to config file")
	etcdServers        = flag.String("etcdServers", "192.168.1.232:2379", "Comma Separated list of etcd servers")
	debug              = flag.Bool("debug", false, "Debug mode")
	cqEnabled          = flag.Bool("cq", false, "Command Queue Enabled")
	newDeviceIsPerson  = flag.Bool("newDeviceIsPerson", false, "Track new devices as people")
)

var bleDevices = []*pb.BleDevices{}

var (
	syncStatusWithGA = time.Hour.Seconds()
	//metrics          map[string]*prometheus.GaugeVec

	devicesPrefix       = "/devices/"
	HomePrefix          = "/homes/"
	AlivePrefix         = "/alive/"
	BlesPrefix          = "/bles/"
	tcPrefix            = "/cq/"
	peoplePrefix        = "/people/"
	notificationsPrefix = "/notifications/"
)

const meterName = "github.com/beaujr/nmap_prometheus"

var devices, lastseen, distance, bledistance, cq api.Float64ObservableGauge
var grpc, grpcEndpoint api.Int64Counter
var grpcAgentEndpoint api.Int64ObservableGauge
var meter api.Meter

func init() {
	exporter, err := prometheus.New()
	if err != nil {
		log.Fatal(err)
	}
	provider := metric.NewMeterProvider(
		metric.WithReader(exporter),
	)
	meter = provider.Meter(meterName)
	devices, err = meter.Float64ObservableGauge("home_detector_device", api.WithDescription("Device in home"))
	if err != nil {
		log.Fatal(err)
	}
	lastseen, err = meter.Float64ObservableGauge("home_detector_device_lastseen", api.WithDescription("Device in home"))
	if err != nil {
		log.Fatal(err)
	}
	distance, err = meter.Float64ObservableGauge("home_detector_device_distance", api.WithDescription("Device in home"))
	if err != nil {
		log.Fatal(err)
	}
	bledistance, err = meter.Float64ObservableGauge("home_detector_ble_distance", api.WithDescription("Device in home"))
	if err != nil {
		log.Fatal(err)
	}

	grpc, err = meter.Int64Counter("grpc_address_count", api.WithDescription("Device in home"))
	if err != nil {
		log.Fatal(err)
	}
	grpcEndpoint, err = meter.Int64Counter("home_detector_grpc_endpoint", api.WithDescription("Device in home"))
	if err != nil {
		log.Fatal(err)
	}

	grpcAgentEndpoint, err = meter.Int64ObservableGauge("home_detector_grpc_clients", api.WithDescription("Device in home"))
	if err != nil {
		log.Fatal(err)
	}
	cq, err = meter.Float64ObservableGauge("home_detector_ble_device", api.WithDescription("Device in home"))
	if err != nil {
		log.Fatal(err)
	}
}

//
//var (
//	peopleHome = promauto.NewGauge(prometheus.GaugeOpts{
//		Name: "home_detector_people_home",
//		Help: "The total number of houseDevices at home",
//	})
//)

// HomeManager manages devices and metric collection
type HomeManager interface {
	deviceDetectState(phone int64) int64
	deviceManager(ctx context.Context) error
	isDeviceOn(iot *pb.Devices) (bool, error)
	IsHouseEmpty(ctx context.Context, home string) bool
	httpHealthCheck(url string) bool
	loadMetrics()
	Devices(w http.ResponseWriter, req *http.Request)
	People(w http.ResponseWriter, req *http.Request)
	HomeEmptyState(w http.ResponseWriter, req *http.Request)
	GetContext() context.Context
}

// Server is an implementation of the proto HomeDetectorServer
type Server struct {
	pb.UnimplementedHomeDetectorServer
	Kv                 etcdv3.KV
	AssistantClient    GoogleAssistant
	EtcdClient         Leaser
	NotificationClient Notifier
	ctx                context.Context
	Logger             *slog.Logger
}

func (s *Server) deviceManager(ctx context.Context) error {
	return nil
}

type Leaser interface {
	GrantLease(ctx context.Context, path, mac string, ttl int64) (string, *etcdv3.LeaseID, error)
	DeleteLeaseByKey(ctx context.Context, key string) error
	GetLeaseByKey(ctx context.Context, key string) (*etcdv3.LeaseStatus, *etcdv3.LeaseTimeToLiveResponse, error)
}

type EtcdLeaser struct {
	etcdv3.Lease
}

func (leaser *EtcdLeaser) GetLeaseByKey(ctx context.Context, key string) (*etcdv3.LeaseStatus, *etcdv3.LeaseTimeToLiveResponse, error) {
	leases, err := leaser.Leases(ctx)
	if err != nil {
		return nil, nil, err
	}
	for _, lease := range leases.Leases {
		leaseTTL, err := leaser.TimeToLive(ctx, lease.ID, etcdv3.WithAttachedKeys())
		if err != nil {
			return nil, nil, err
		}
		for _, leasedItem := range leaseTTL.Keys {
			if strings.Compare(string(leasedItem), key) == 0 {
				return &lease, leaseTTL, nil
			}
		}
	}
	return nil, nil, nil
}

func (leaser *EtcdLeaser) DeleteLeaseByKey(ctx context.Context, key string) error {
	leases, err := leaser.Leases(ctx)
	if err != nil {
		return err
	}
	for _, lease := range leases.Leases {
		leaseTTL, err := leaser.TimeToLive(ctx, lease.ID, etcdv3.WithAttachedKeys())
		if err != nil {
			return err
		}
		for _, leasedItem := range leaseTTL.Keys {
			if strings.Compare(string(leasedItem), key) == 0 {
				_, err = leaser.Revoke(ctx, lease.ID)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func NewEtcdLeaser(lease etcdv3.Lease) Leaser {
	return &EtcdLeaser{lease}
}

func (leaser *EtcdLeaser) GrantLease(ctx context.Context, path, mac string, ttl int64) (string, *etcdv3.LeaseID, error) {
	keyPath := filepath.Join(AlivePrefix, path, mac)
	leaseExists := false
	leases, err := leaser.Leases(ctx)
	if err != nil {
		return "", nil, err
	}
	for _, lease := range leases.Leases {
		leaseTTL, err := leaser.TimeToLive(ctx, lease.ID, etcdv3.WithAttachedKeys())
		if err != nil {
			return "", nil, err
		}
		for _, key := range leaseTTL.Keys {
			if strings.Compare(keyPath, string(key)) == 0 {
				leaseExists = true
				_, err := leaser.KeepAliveOnce(ctx, leaseTTL.ID)
				if err != nil {
					return "", nil, err
				}

			}
		}
		if leaseExists {
			break
		}
	}
	if !leaseExists {
		lease, err := leaser.Grant(ctx, ttl)
		if err != nil {
			return "", nil, err
		}
		if lease != nil {
			return keyPath, &lease.ID, err
		}
	}
	return "", nil, nil
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

// NewCustomServer function to allow passing in Server dependencies
func NewCustomServer(ctx context.Context, e etcdv3.KV, g GoogleAssistant, n Notifier, handler slog.Handler) Server {
	s := Server{
		UnimplementedHomeDetectorServer: pb.UnimplementedHomeDetectorServer{},
		Kv:                              e,
		AssistantClient:                 g,
		NotificationClient:              n,
		ctx:                             ctx,
		Logger:                          slog.New(handler),
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
	//		s.Logger.Info(err)
	//	}
	//	err = server.iotStatusManager()
	//	if err != nil {
	//		s.Logger.Info(err)
	//	}
	//})
	if *cqEnabled {
		c.AddFunc("*/10 * * * * *", func() {
			err := server.processTimedCommandQueue()
			if err != nil {
				server.Logger.Error(err.Error())
			}
		})
	}
	c.Start()
	return
}

// NewServer new instance of HomeManager
func NewServer(ctx context.Context) HomeManager {
	client, etcdClient := etcd.NewClient(strings.Split(*etcdServers, ","))
	assistantClient := NewAssistant()
	notifyClient := NewNotifier(etcdClient)
	server := &Server{Kv: etcdClient, AssistantClient: assistantClient, NotificationClient: notifyClient, EtcdClient: NewEtcdLeaser(client.Lease), ctx: ctx, Logger: slog.New(slog.NewTextHandler(os.Stderr, nil))}
	_, err := server.ReadNetworkConfig()
	if err != nil {
		server.Logger.Error(err.Error())
	}
	// importConfig to etcd
	bleDevices, _ := readBleConfig(*bleConfigFile)
	for _, item := range bleDevices {
		_ = server.writeBleDevice(item)
	}

	// Bluetooth
	bles, err := server.ReadBleConfig()
	if err != nil {
		server.Logger.Error(err.Error())
	}
	for _, item := range bles {
		server.RegisterBleMetric(item, "etcd")
	}

	knowDevices, err := server.ReadNetworkConfig()
	if err != nil {
		server.Logger.Error(err.Error())
	}
	homes := make([]string, 0)
	for _, item := range knowDevices {
		homes = append(homes, item.Home)
	}

	for _, home := range homes {
		homeKey := fmt.Sprintf("%s%s", HomePrefix, home)
		_, err := server.Kv.Put(ctx, homeKey, strconv.FormatBool(server.IsHouseEmpty(ctx, home)))
		if err != nil {
			server.Logger.Error(err.Error())
		}
	}
	createCrons(server)
	server.loadMetrics()
	return server
}

func (s *Server) GetContext() context.Context {
	return s.ctx
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

type networkDevice struct {
	*pb.Devices
}

func (d *networkDevice) observe(_ context.Context, obs api.Observer) error {
	away := float64(1)
	if (time.Now().Unix() - d.GetLastSeen()) > *TimeAwaySeconds {
		away = float64(0)
	}
	attrs := []attribute.KeyValue{
		attribute.Key("name").String(d.GetName()),
		attribute.Key("mac").String(d.GetId().GetMac()),
		attribute.Key("home").String(d.GetHome()),
		attribute.Key("person").Bool(d.GetPerson()),
	}
	obs.ObserveFloat64(devices, away, api.WithAttributes(attrs...))
	obs.ObserveFloat64(lastseen, float64(d.GetLastSeen()), api.WithAttributes(attrs...))
	obs.ObserveFloat64(distance, float64(d.GetLatency()), api.WithAttributes(attrs...))
	return nil
}

type bluetoothDevice struct {
	*pb.BleDevices
	agent string
}

func (d *bluetoothDevice) observe(ctx context.Context, obs api.Observer) error {
	attrs := []attribute.KeyValue{
		attribute.Key("name").String(d.GetName()),
		attribute.Key("mac").String(d.GetId()),
		attribute.Key("home").String(d.GetHome()),
	}
	obs.ObserveFloat64(lastseen, float64(d.GetLastSeen()), api.WithAttributes(append(attrs, attribute.Key("person").Bool(false))...))
	obs.ObserveFloat64(bledistance, float64(d.GetDistance()), api.WithAttributes(append(attrs, attribute.Key("agent").String(d.agent))...))
	return nil
}
func (s *Server) RegisterMetric(item *pb.Devices) {
	d := &networkDevice{item}
	_, err := meter.RegisterCallback(d.observe, lastseen, distance, devices)
	if err != nil {
		log.Panicln(err.Error())
	}
}

func (s *Server) RegisterBleMetric(item *pb.BleDevices, agent string) {
	b := &bluetoothDevice{item, agent}
	_, err := meter.RegisterCallback(b.observe, bledistance, lastseen)
	if err != nil {
		log.Panicln(err.Error())
	}
}

func (s *Server) loadMetrics() {
	knowDevices, err := s.ReadNetworkConfig()
	if err != nil {
		s.Logger.Info(err.Error())
	}
	for _, item := range knowDevices {
		s.RegisterMetric(item)
	}
	// Bluetooth
	bles, err := s.ReadBleConfig()
	if err != nil {
		s.Logger.Error(err.Error())
	}
	for _, item := range bles {
		s.RegisterBleMetric(item, "etcd")
	}
}

//func (s *Server) RecordLastSeenMetric(item *pb.Devices, timestamp int64) {
//	metrics["lastseen"].WithLabelValues(item.Name, item.GetId().GetMac(), item.GetHome(), strconv.FormatBool(item.GetPerson())).Set(float64(timestamp))
//	timeAway := s.deviceDetectState(item.LastSeen)
//	if timeAway > *TimeAwaySeconds {
//		metrics["hdd"].WithLabelValues(item.Name, item.GetId().GetMac(), item.GetHome(), strconv.FormatBool(item.GetPerson())).Set(float64(0))
//	} else {
//		metrics["hdd"].WithLabelValues(item.Name, item.GetId().GetMac(), item.GetHome(), strconv.FormatBool(item.GetPerson())).Set(float64(1))
//	}
//}
//
//func (s *Server) RecordDistanceMetric(item *pb.Devices, distance float32) {
//	metrics["distance"].WithLabelValues(item.Name, item.GetId().GetMac(), item.GetHome(), strconv.FormatBool(item.GetPerson())).Set(float64(distance))
//}

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
	s.Logger.Info(fmt.Sprintf("New BLE Device: %s", in.Key))

	bleDevices = append(bleDevices, &newDevice)
	_, err := uniqueBle(bleDevices)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) newDevice(ctx context.Context, in *pb.AddressRequest, home string, md []*pb.Metadata) error {
	name := in.Ip

	if in.Mac != "" && len(in.GetHosts()) == 0 {
		name = in.Mac
	}

	if len(in.GetHosts()) > 0 {
		name = in.GetHosts()[0]
	}
	vendor := "unknown"
	name = strings.ReplaceAll(strings.ReplaceAll(name, ".", "_"), ":", "_")
	if in.GetVendor() != "" {
		vendor = in.GetVendor()
		name = vendor
	} else if in.Mac != in.Ip && strings.Contains(in.Mac, ":") {
		macVendor, err := GetManufacturer(in.Mac)
		if macVendor != nil {
			vendor = *macVendor
			name = vendor
		}
		if err != nil {
			s.Logger.Error(err.Error())
			vendor = name
		}
	}
	newDevice := pb.Devices{
		Name:         name,
		Id:           &pb.NetworkId{Ip: in.Ip, Mac: in.Mac, UUID: in.Mac},
		Away:         false,
		LastSeen:     int64(time.Now().Unix()),
		Person:       *newDeviceIsPerson,
		Command:      "",
		Manufacturer: vendor,
		Home:         home,
		Hostnames:    in.Hosts,
		Metadata:     md,
	}
	if len(in.Hosts) > 0 {
		newDevice.Name = in.Hosts[0]
	}
	s.Logger.Info(fmt.Sprintf("New Device: %s", name))

	err := s.WriteNetworkDevice(ctx, &newDevice)
	if err != nil {
		s.Logger.Info("Error saving to ETCD: %s", err.Error())
	}
	err = s.NotificationClient.SendNotification(fmt.Sprintf("New Device in %s (%s)", newDevice.Home, newDevice.Id.Ip), fmt.Sprintf("%s (%s)", newDevice.Name, newDevice.Manufacturer), newDevice.Home)
	if err != nil {
		s.Logger.Info("Error sending notification: %s", err.Error())
	}
	s.RegisterMetric(&newDevice)
	return nil
}

func (s *Server) existingDevice(ctx context.Context, houseDevice *pb.Devices, incoming *pb.AddressRequest, home string) error {
	if incoming.GetVendor() != "" && incoming.GetVendor() != houseDevice.GetManufacturer() {
		houseDevice.Manufacturer = incoming.GetVendor()
	}
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
	// this is all handled via ttl now and leases
	//if houseDevice.GetPerson() {
	//	err := s.processPerson(houseDevice)
	//	if err != nil {
	//		return nil
	//	}
	//}
	houseDevice.LastSeen = int64(time.Now().Unix())
	houseDevice.Latency = incoming.GetDistance()
	if len(incoming.Hosts) > 0 {
		hostnamesMaps := map[string]bool{}
		//load existing
		for _, host := range houseDevice.Hostnames {
			hostnamesMaps[host] = true
		}
		for _, host := range incoming.Hosts {
			if _, exist := hostnamesMaps[host]; !exist {
				houseDevice.Hostnames = append(houseDevice.Hostnames, host)
			}
		}
	}
	if incoming.Mac != "" && incoming.Mac == houseDevice.Id.Mac {
		err := s.WriteNetworkDevice(ctx, houseDevice)
		if err != nil {
			s.Logger.Info("Error saving to ETCD: %s", err.Error())
		}
		s.RegisterMetric(houseDevice)
	}
	return nil
}

func (s *Server) GrantLease(ctx context.Context, data map[string]string, ttl int64) error {
	key, leaseId, err := s.EtcdClient.GrantLease(ctx, data["home"], data["mac"], ttl)
	if err != nil {
		return err
	}

	if leaseId != nil && key != "" {
		_, err = s.Kv.Put(ctx, key, data["value"], etcdv3.WithLease(*leaseId))
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

func (s *Server) ListPeopleRequest(ctx context.Context) (*pb.PeopleResponse, error) {
	var people []*pb.People
	resp, err := s.Kv.Get(ctx, peoplePrefix, etcdv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	for _, person := range resp.Kvs {
		var human pb.People
		err := yaml.Unmarshal(person.Value, &human)
		if err != nil {
			return nil, err
		}
		for _, device := range human.GetIds() {
			resp, err := s.Kv.Get(ctx, fmt.Sprintf("%s%s", AlivePrefix, device))
			if err != nil {
				return nil, err
			}
			if len(resp.Kvs) > 0 {
				for _, deviceId := range resp.Kvs {
					human.Home = string(deviceId.Value)
					human.Away = false

				}
			}
		}
		people = append(people, &human)
	}
	return &pb.PeopleResponse{People: people}, nil
}

func (s *Server) ProcessIncomingAddress(ctx context.Context, in *pb.AddressRequest) (*pb.Reply, error) {
	incoming := in
	headers, _ := metadata.FromIncomingContext(ctx)
	home := "unknown"
	val := headers.Get("home")
	if len(val) > 0 {
		home = val[0]
	}

	typeOfDevice := agent.NetworkType
	deviceType := headers.Get("type")
	if len(deviceType) > 0 {
		typeOfDevice = deviceType[0]
	}
	md := []*pb.Metadata{{Key: "type", Value: typeOfDevice}}

	if incoming.Mac == "" && home != "" {
		incoming.Mac = fmt.Sprintf("%s/%s", home, strings.ReplaceAll(in.Ip, ".", "_"))
	}
	opts := []etcdv3.OpOption{
		etcdv3.WithLimit(1),
	}
	opts = append(opts, etcdv3.WithLastRev()...)

	item, err := s.Kv.Get(ctx, fmt.Sprintf("%s%s", devicesPrefix, in.Mac), opts...)
	if err != nil {
		return nil, err
	}
	path := "device"
	if item.Count == 0 {
		if *newDeviceIsPerson {
			path = "person"
		}
		err := s.newDevice(ctx, in, home, md)
		if err != nil {
			return nil, err
		}
		// grant lease after update for new person
		err = s.GrantLease(ctx, map[string]string{"mac": in.Mac, "home": home, "value": path}, *TimeAwaySeconds)
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
		if received := ctx.Value("received"); received != nil {
			iReceived := received.(int64)
			if exDevice.GetLastSeen() > iReceived {
				return &pb.Reply{Acknowledged: true}, nil
			}
		}
		exDevice.Metadata = md
		if exDevice.GetPerson() {
			path = "person"
		}
		// grant lease before update for existing person
		err = s.GrantLease(ctx, map[string]string{"mac": in.Mac, "home": home, "value": path}, *TimeAwaySeconds)
		if err != nil {
			return nil, err
		}
		err = s.existingDevice(ctx, exDevice, incoming, home)
		if err != nil {
			return nil, err
		}
	}
	return &pb.Reply{Acknowledged: true}, nil
}
func (s *Server) grpcHitsMetrics(ctx context.Context, name string, itemCount int) {
	attrs := []attribute.KeyValue{
		attribute.Key("name").String(name),
	}
	grpc.Add(ctx, int64(itemCount), api.WithAttributes(attrs...))
	md, _ := metadata.FromIncomingContext(ctx)
	_, err := meter.RegisterCallback(headers{md}.observeGrpc, grpcAgentEndpoint)
	if err != nil {
		log.Panicln(err.Error())
	}
}

type headers struct {
	metadata.MD
}

func (md headers) observeGrpc(_ context.Context, obs api.Observer) error {
	attrs := []attribute.KeyValue{}
	val := md.Get("home")
	if len(val) > 0 {
		attrs = append(attrs, attribute.Key("home").String(val[0]))
	}
	agentType := md.Get("type")
	if len(agentType) > 0 {
		attrs = append(attrs, attribute.Key("type").String(agentType[0]))
	}
	if val := md.Get("client"); len(val) > 0 {
		attrs = append(attrs, attribute.Key("name").String(val[0]))
	}
	obs.ObserveInt64(grpcAgentEndpoint, time.Now().Unix(), api.WithAttributes(attrs...))
	return nil
}

func (s *Server) grpcPrometheusMetrics(ctx context.Context, promMetric string, name string) {
	grpcEndpoint.Add(ctx, 1, api.WithAttributes([]attribute.KeyValue{
		attribute.Key("name").String(name)}...))
}

func (s *Server) getBLEById(id *string) (*pb.BleDevices, error) {
	opts := []etcdv3.OpOption{
		etcdv3.WithLimit(1),
	}
	s.Logger.Info(fmt.Sprintf("%s%s", BlesPrefix, *id))
	item, err := s.Kv.Get(s.GetContext(), fmt.Sprintf("%s%s", BlesPrefix, *id), opts...)
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
	headers, _ := metadata.FromIncomingContext(ctx)
	md := []string{}
	for _, item := range device.GetMetadata() {
		if !slices.Contains(md, item.GetKey()) {
			md = append(md, item.GetKey())
		}
	}
	for k, v := range headers {
		if !slices.Contains(md, k) {
			md = append(md, k)
			if len(v) > 0 {
				device.Metadata = append(device.Metadata, &pb.Metadata{Key: k, Value: strings.Join(v, "|")})
			}
		}
	}
	device.LastSeen = time.Now().Unix()
	device.Distance = in.Distance

	err = s.writeBleDevice(device)
	if err != nil {
		s.Logger.Info(fmt.Sprintf("Error updating BLE device: %s", device.Id))
	}
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

func (s *Server) IsHouseEmpty(ctx context.Context, home string) bool {
	peopleAtHome, err := s.GetPeopleInHouses(ctx, home)
	if err != nil {
		s.Logger.Info("Failed to read from etcd")
	}
	return len(peopleAtHome) == 0
}
