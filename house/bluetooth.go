package house

import (
	"context"
	"fmt"
	"github.com/ozonru/etcd/v3/clientv3"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type bleDevice struct {
	Id       string    `json:"id",yaml:"id"`
	LastSeen int64     `json:"lastSeen",yaml:"lastSeen"`
	Commands []command `json:"commands",yaml:"commands"`
	Name     string    `json:"name",yaml:"name"`
	Home     string    `json:"home",yaml:"home"`
}

type command struct {
	Timeout        int64  `json:"timeout",yaml:"timeout"`
	Command        string `json:"command",yaml:"command"`
	TimeoutCommand string `json:"timeoutcommand",yaml:"timeoutcommand"`
}

// TimedCommand executes a command now and a reverse command in now + executeat seconds
type TimedCommand struct {
	Owner     string `json:"mac",yaml:"mac"`
	Command   string `json:"command",yaml:"command"`
	ExecuteAt int64  `json:"executeat",yaml:"executeat"`
	Executed  bool   `json:"executed",yaml:"executed"`
}

func writeBleDevices(devices []*bleDevice) error {
	d1, err := yaml.Marshal(devices)
	if err != nil {
		return err
	}
	return writeConfig(d1, *bleConfigFile)
}

func readBleConfig(filename string) ([]*bleDevice, error) {
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

	var result []*bleDevice
	err = yaml.Unmarshal(byteValue, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Server) readBleConfig() (map[string]*bleDevice, error) {
	var result map[string]*bleDevice
	result = make(map[string]*bleDevice)
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
		var dev *bleDevice
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

func uniqueBle(devices []*bleDevice) ([]*bleDevice, error) {
	keys := make(map[string]bool)
	list := []*bleDevice{}
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
func (s *Server) writeBleDevice(item *bleDevice) error {
	d1, err := yaml.Marshal(item)
	if err != nil {
		log.Fatalf(err.Error())
	}

	key := fmt.Sprintf("%s%s", blesPrefix, item.Id)
	_, err = s.etcdClient.Put(context.Background(), key, string(d1))
	return err
}
