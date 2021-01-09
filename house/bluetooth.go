package house

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

type bleDevice struct {
	Id       string    `json:"id",yaml:"id"`
	LastSeen int64     `json:"lastSeen",yaml:"lastSeen"`
	Commands []command `json:"commands",yaml:"commands"`
	Name     string    `json:"name",yaml:"name"`
}

type command struct {
	Timeout        int64  `json:"timeout",yaml:"timeout"`
	Command        string `json:"command",yaml:"command"`
	TimeoutCommand string `json:"timeoutcommand",yaml:"timeoutcommand"`
}

// TimedCommand executes a command now and a reverse command in now + executeat seconds
type TimedCommand struct {
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
