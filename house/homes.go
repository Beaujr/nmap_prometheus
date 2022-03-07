package house

import (
	"context"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"strconv"
	"strings"
)

var (
	houseTimeOut = flag.Int64("absence", 3600, "How long a house is empty (in seconds) before turning off smart devices.")
)

type home struct {
	Empty   bool `json:"empty",yaml:"empty"`
	Timeout int  `json:"timeout",yaml:"timeout"`
}

func (s *Server) writeHome(id string, item *home) error {
	d1, err := yaml.Marshal(item)
	if err != nil {
		log.Fatalf(err.Error())
	}

	key := fmt.Sprintf("%s%s", homePrefix, id)
	_, err = s.EtcdClient.Put(context.Background(), key, string(d1))
	return err
}

func (s *Server) iotStatusManager() error {
	gHouseEmpty, err := s.ReadHomesConfig()
	if err != nil {
		return err
	}
	for home, empty := range gHouseEmpty {
		if houseEmpty := s.IsHouseEmpty(home); houseEmpty != *empty {
			err = s.ToggleHouseStatus(home, houseEmpty)
			if err != nil {
				return err
			}
		}
		if !s.IsHouseEmpty(home) {
			keys, err := s.getTcKeys()
			if err != nil {
				return err
			}
			for _, key := range keys {
				if strings.Contains(*key, home) {
					err = s.deleteTcByKey(*key)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (s *Server) ToggleHouseStatus(home string, houseEmpty bool) error {
	_, err := s.EtcdClient.Put(context.Background(), fmt.Sprintf("%s%s", homePrefix, home), strconv.FormatBool(houseEmpty))
	if err != nil {
		log.Println(err)
		return err
	}

	err = s.NotificationClient.SendNotification("House Empty", fmt.Sprintf("No Humans in %s", home), home)
	if err != nil {
		log.Println(err)
		return err
	}

	devices, err := s.ReadNetworkConfig()
	if err != nil {
		log.Println(err)
		return err
	}
	i := int64(0)
	for _, device := range devices {
		if device.PresenceAware && strings.Compare(home, device.Home) == 0 {
			err = s.createTimedCommand(*houseTimeOut+(10*i), device.Id.Mac, home, fmt.Sprintf("Turn %s off", device.Name), device.Home)
			if err != nil {
				return err
			}
			i++
		}
	}
	return err
}
