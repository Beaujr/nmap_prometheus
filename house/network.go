package house

import (
	"context"
	"fmt"
	"github.com/ozonru/etcd/v3/clientv3"
	"gopkg.in/yaml.v2"
	"log"
	"strings"
)

type device struct {
	Id                 networkId `json:"id",yaml:"id"`
	Home               string    `json:"home",yaml:"home"`
	LastSeen           int64     `json:"lastSeen",yaml:"lastSeen"`
	Away               bool      `json:"away",yaml:"away"`
	Name               string    `json:"name",yaml:"name"`
	Person             bool      `json:"person",yaml:"person"`
	Command            string    `json:"command",yaml:"command"`
	Smart              bool      `json:"smart",yaml:"smart"`
	Manufacturer       string    `json:"manufacturer",yaml:"manufacturer"`
	SmartStatusCommand string    `json:"gaStatusCmd,omitempty",yaml:"gaStatusCmd,omitempty"`
	PresenceAware      bool      `json:"aware,omitempty",yaml:"aware,omitempty"`
}

func writeNetworkDevices(devices map[string]*device) error {
	d1, err := yaml.Marshal(devices)
	if err != nil {
		return err
	}
	return writeConfig(d1, *networkConfigFile)
}

func (s *Server) writeNetworkDevice(item *device) error {
	d1, err := yaml.Marshal(item)
	if err != nil {
		log.Fatalf(err.Error())
	}

	key := fmt.Sprintf("%s/%s", devicesPrefix, item.Id.UUID)
	_, err = s.etcdClient.Put(context.Background(), key, string(d1))
	return err
}

func (s *Server) readNetworkConfig() (map[string]*device, error) {
	var result map[string]*device
	result = make(map[string]*device)
	items, err := s.etcdClient.Get(context.Background(), "", clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if items == nil {
		return result, nil
	}
	i := 0
	for i < int(items.Count) {
		val := items.Kvs[i].Value
		key := items.Kvs[i].Key
		// Once off Beau Code
		//if !strings.Contains(string(key), devicesPrefix) {
		//	_, err := s.etcdClient.Delete(context.Background(), string(key))
		//	if err != nil {
		//		log.Fatalf(err.Error())
		//	}
		//}
		//// end once off
		var dev *device
		err = yaml.Unmarshal(val, &dev)
		if err != nil {
			return nil, err
		}
		newKey := strings.ReplaceAll(string(key), "/devices/", "")
		result[string(newKey)] = dev
		i++
	}
	return result, nil
}
