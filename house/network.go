package house

import (
	"context"
	"fmt"
	pb "github.com/beaujr/nmap_prometheus/proto"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

func readDevicesConfig(filename string) ([]*pb.Devices, error) {
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

	var result []*pb.Devices
	err = yaml.Unmarshal(byteValue, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func writeNetworkDevices(devices map[string]*pb.Devices) error {
	d1, err := yaml.Marshal(devices)
	if err != nil {
		return err
	}
	return writeConfig(d1, *networkConfigFile)
}

func (s *Server) WriteNetworkDevice(ctx context.Context, item *pb.Devices) error {
	d1, err := yaml.Marshal(item)
	if err != nil {
		log.Fatalf(err.Error())
	}

	key := fmt.Sprintf("%s%s", devicesPrefix, item.Id.UUID)
	_, err = s.Kv.Put(ctx, key, string(d1))
	return err
}
func (s *Server) ReadNetworkConfig() (map[string]*pb.Devices, error) {
	var result map[string]*pb.Devices
	result = make(map[string]*pb.Devices)
	items, err := s.Kv.Get(s.GetContext(), devicesPrefix, clientv3.WithPrefix())
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

func (s *Server) getDevices(ctx context.Context) ([]*pb.Devices, error) {
	var result []*pb.Devices
	result = make([]*pb.Devices, 0)
	//leadCtx := clientv3.WithRequireLeader(ctx)
	items, err := s.Kv.Get(ctx, devicesPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if items == nil {
		return nil, nil
	}
	i := 0
	for i < int(items.Count) {
		val := items.Kvs[i].Value
		var dev *pb.Devices
		err = yaml.Unmarshal(val, &dev)
		if err != nil {
			return nil, err
		}
		result = append(result, dev)
		i++
	}
	return result, nil
}

func (s *Server) GetDevice(id string) (*pb.Devices, error) {
	items, err := s.Kv.Get(s.GetContext(), fmt.Sprintf("%s%s", devicesPrefix, id))
	if err != nil {
		return nil, err
	}
	if items == nil {
		return nil, nil
	}
	if items.Count != 1 {
		return nil, fmt.Errorf("coulnt find distinct item for: %s", id)
	}
	val := items.Kvs[0].Value
	var dev *pb.Devices
	err = yaml.Unmarshal(val, &dev)
	if err != nil {
		return nil, err
	}
	return dev, nil
}

func (s *Server) deleteDeviceById(id string) error {
	_, err := s.Kv.Delete(s.GetContext(), fmt.Sprintf("%s%s", devicesPrefix, id))
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) processPerson(houseDevice *pb.Devices) error {
	homeKey := fmt.Sprintf("%s%s", HomePrefix, houseDevice.Home)
	houseStatus, err := s.Kv.Get(s.GetContext(), homeKey)
	if err != nil {
		log.Panic(err.Error())
	}

	if houseStatus.Count == 0 {
		homeKey := fmt.Sprintf("%s%s", HomePrefix, houseDevice.Home)
		_, err = s.Kv.Put(s.GetContext(), homeKey, "false")
		if err != nil {
			log.Panic(err.Error())
		}
	} else if val, err := strconv.ParseBool(string(houseStatus.Kvs[0].Value)); val && err == nil {
		homeKey := fmt.Sprintf("%s%s", HomePrefix, houseDevice.Home)
		_, err = s.Kv.Put(s.GetContext(), homeKey, "false")
		if err != nil {
			log.Panic(err.Error())
		}

		err := s.NotificationClient.SendNotification(houseDevice.Home, "No longer Empty", houseDevice.Home)
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

func (s *Server) ReadHomesConfig() (map[string]*bool, error) {
	var result map[string]*bool
	result = make(map[string]*bool)
	items, err := s.Kv.Get(s.GetContext(), HomePrefix, clientv3.WithPrefix())
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
		newKey := strings.ReplaceAll(string(key), HomePrefix, "")
		boolVal, _ := strconv.ParseBool(string(val))
		if strings.Contains(string(key), "//") {
			key2 := strings.ReplaceAll(string(key), "//", "")
			_, err := s.Kv.Put(s.GetContext(), key2, string(val))
			if err != nil {
				return nil, err
			}
			_, err = s.Kv.Delete(s.GetContext(), string(key))
			if err != nil {
				return nil, err
			}
		}
		result[string(newKey)] = &boolVal
		i++
	}
	return result, nil
}
