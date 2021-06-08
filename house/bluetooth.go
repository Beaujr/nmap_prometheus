package house

import (
	"context"
	"fmt"
	pb "github.com/beaujr/nmap_prometheus/proto"
	"github.com/ozonru/etcd/v3/clientv3"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func writeBleDevices(devices []*pb.BleDevices) error {
	d1, err := yaml.Marshal(devices)
	if err != nil {
		return err
	}
	return writeConfig(d1, *bleConfigFile)
}

func readBleConfig(filename string) ([]*pb.BleDevices, error) {
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

	var result []*pb.BleDevices
	err = yaml.Unmarshal(byteValue, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Server) readBleConfig() (map[string]*pb.BleDevices, error) {
	var result map[string]*pb.BleDevices
	result = make(map[string]*pb.BleDevices)
	items, err := s.etcdClient.Get(context.Background(), blesPrefix, clientv3.WithPrefix())
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
		var dev *pb.BleDevices
		err = yaml.Unmarshal(val, &dev)
		if err != nil {
			return nil, err
		}
		strKey := string(key)
		newKey := strings.ReplaceAll(strKey, blesPrefix, "")
		result[string(newKey)] = dev
		i++
	}
	return result, nil
}

func (s *Server) readBleConfigAsSlice() ([]*pb.BleDevices, error) {
	var result []*pb.BleDevices
	result = make([]*pb.BleDevices, 0)
	items, err := s.etcdClient.Get(context.Background(), blesPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if items == nil {
		return result, nil
	}
	i := 0
	for i < int(items.Count) {
		val := items.Kvs[i].Value
		var dev *pb.BleDevices
		err = yaml.Unmarshal(val, &dev)
		if err != nil {
			return nil, err
		}
		result = append(result, dev)
		i++
	}
	return result, nil
}

func uniqueBle(devices []*pb.BleDevices) ([]*pb.BleDevices, error) {
	keys := make(map[string]bool)
	list := []*pb.BleDevices{}
	for _, entry := range devices {
		if _, value := keys[entry.Id]; !value {
			keys[entry.Id] = true
			list = append(list, entry)
		}
	}
	err := writeBleDevices(list)
	if err != nil {
		return nil, err
	}
	return list, nil
}
func (s *Server) writeBleDevice(item *pb.BleDevices) error {
	d1, err := yaml.Marshal(item)
	if err != nil {
		log.Fatalf(err.Error())
	}

	key := fmt.Sprintf("%s%s", blesPrefix, item.Id)
	_, err = s.etcdClient.Put(context.Background(), key, string(d1))
	return err
}

func (s *Server) writeTc(item *pb.TimedCommands) error {
	d1, err := yaml.Marshal(item)
	if err != nil {
		log.Fatalf(err.Error())
	}

	key := fmt.Sprintf("%s%s", tcPrefix, item.Id)
	_, err = s.etcdClient.Put(context.Background(), key, string(d1))
	return err
}

func (s *Server) deleteTc(item *pb.TimedCommands) error {
	key := fmt.Sprintf("%s%s", tcPrefix, item.Id)
	return s.deleteTcByKey(key)
}

func (s *Server) deleteTcByKey(key string) error {
	_, err := s.etcdClient.Delete(context.Background(), key)
	return err
}

func (s *Server) getTc() (map[string]*pb.TimedCommands, error) {
	var result map[string]*pb.TimedCommands
	result = make(map[string]*pb.TimedCommands)
	items, err := s.etcdClient.Get(context.Background(), tcPrefix, clientv3.WithPrefix())
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
		var dev *pb.TimedCommands
		err = yaml.Unmarshal(val, &dev)
		if err != nil {
			return nil, err
		}
		strKey := string(key)
		newKey := strings.ReplaceAll(strKey, tcPrefix, "")
		result[string(newKey)] = dev
		i++
	}
	return result, nil
}

func (s *Server) getTcById(id string) (*pb.TimedCommands, error) {
	items, err := s.etcdClient.Get(context.Background(), fmt.Sprintf("%s%s", tcPrefix, id), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if items.Count == 0 {
		return nil, fmt.Errorf("CQ with id:%s not found", id)
	}
	var dev *pb.TimedCommands
	err = yaml.Unmarshal(items.Kvs[0].Value, &dev)
	if err != nil {
		return nil, err
	}
	return dev, nil
}

func (s *Server) getTcKeys() ([]*string, error) {
	var result []*string
	result = make([]*string, 0)
	items, err := s.etcdClient.Get(context.Background(), tcPrefix, clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if err != nil {
		return nil, err
	}
	if items == nil {
		return result, nil
	}
	i := 0
	for i < int(items.Count) {
		strKey := string(items.Kvs[i].Key)
		result = append(result, &strKey)
		i++
	}
	return result, nil
}
