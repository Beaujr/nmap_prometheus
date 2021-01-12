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

func writeNetworkDevices(devices map[string]*device) error {
	d1, err := yaml.Marshal(devices)
	if err != nil {
		return err
	}
	return writeConfig(d1, *networkConfigFile)
}

func readNetworkConfig(filename string) (map[string]*device, error) {
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

	var result map[string]*device
	err = yaml.Unmarshal(byteValue, &result)
	if err != nil {
		return nil, err
	}
	if result == nil {
		result = make(map[string]*device)
	}
	return result, nil
}
