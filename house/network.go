package house

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
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

func uniqueNetwork(devices []*device) ([]*device, error) {
	keys := make(map[string]bool)
	list := []*device{}
	for _, entry := range devices {
		if _, value := keys[entry.Id.Mac]; !value {
			keys[entry.Id.Mac] = true
			list = append(list, entry)
		}
	}
	err := writeNetworkDevices(list)
	if err != nil {
		return nil, err
	}
	return list, nil
}
func writeNetworkDevices(devices []*device) error {
	d1, err := yaml.Marshal(devices)
	if err != nil {
		return err
	}
	return writeConfig(d1, *networkConfigFile)
}

func readNetworkConfig(filename string) ([]*device, error) {
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

	var result []*device
	err = yaml.Unmarshal(byteValue, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
