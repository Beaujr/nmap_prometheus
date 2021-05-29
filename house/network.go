package house

import (
	"context"
	"fmt"
	pb "github.com/beaujr/nmap_prometheus/proto"
	"github.com/ozonru/etcd/v3/clientv3"
	"gopkg.in/yaml.v2"
	"log"
	"strconv"
	"strings"
)

func writeNetworkDevices(devices map[string]*pb.Devices) error {
	d1, err := yaml.Marshal(devices)
	if err != nil {
		return err
	}
	return writeConfig(d1, *networkConfigFile)
}

func (s *Server) writeNetworkDevice(item *pb.Devices) error {
	d1, err := yaml.Marshal(item)
	if err != nil {
		log.Fatalf(err.Error())
	}

	key := fmt.Sprintf("%s%s", devicesPrefix, item.Id.UUID)
	_, err = s.etcdClient.Put(context.Background(), key, string(d1))
	return err
}
func (s *Server) readNetworkConfig() (map[string]*pb.Devices, error) {
	var result map[string]*pb.Devices
	result = make(map[string]*pb.Devices)
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
		var dev *pb.Devices
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

func (s *Server) processPerson(houseDevice *pb.Devices) error {
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

		err := s.notificationClient.SendNotification(houseDevice.Home, "No longer Empty", houseDevice.Home)
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
