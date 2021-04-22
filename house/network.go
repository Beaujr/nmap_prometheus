package house

import (
	"context"
	"fmt"
	"github.com/ozonru/etcd/v3/clientv3"
	"gopkg.in/yaml.v2"
	"log"
	"strconv"
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

	key := fmt.Sprintf("%s%s", devicesPrefix, item.Id.UUID)
	_, err = s.etcdClient.Put(context.Background(), key, string(d1))
	return err
}
func (s *Server) readNetworkConfig() (map[string]*device, error) {
	var result map[string]*device
	result = make(map[string]*device)
	items, err := s.etcdClient.Get(context.Background(), devicesPrefix, clientv3.WithPrefix())
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
		var dev *device
		err = yaml.Unmarshal(val, &dev)
		if err != nil {
			return nil, err
		}
		strKey := string(key)
		newKey := strings.ReplaceAll(strKey, devicesPrefix, "")
		result[string(newKey)] = dev
		i++
	}
	return result, nil
}

func (s *Server) processPerson(houseDevice *device) error {
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
			err := SendNotification(houseDevice.Home, "No longer Empty", houseDevice.Home)
			if err != nil {
				return err
			}
			tcs, err := s.getTc()
			if err != nil {
				return err
			}
			for key, val := range tcs {
				if strings.Contains(key, houseDevice.Home) {
					err = s.deleteTc(val)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (s *Server) readHomesConfig() (map[string]*bool, error) {
	var result map[string]*bool
	result = make(map[string]*bool)
	items, err := s.etcdClient.Get(context.Background(), homePrefix, clientv3.WithPrefix())
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
		newKey := strings.ReplaceAll(string(key), homePrefix, "")
		boolVal, _ := strconv.ParseBool(string(val))
		if strings.Contains(string(key), "//") {
			key2 := strings.ReplaceAll(string(key), "//", "")
			_, err := s.etcdClient.Put(context.Background(), key2, string(val))
			if err != nil {
				return nil, err
			}
			_, err = s.etcdClient.Delete(context.Background(), string(key))
			if err != nil {
				return nil, err
			}
		}
		result[string(newKey)] = &boolVal
		i++
	}
	return result, nil
}
