package house

import (
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

	key := fmt.Sprintf("%s%s", HomePrefix, id)
	_, err = s.Kv.Put(s.GetContext(), key, string(d1))
	return err
}

func (s *Server) ToggleHouseStatus(home string, houseEmpty bool) error {
	_, err := s.Kv.Put(s.GetContext(), fmt.Sprintf("%s%s", HomePrefix, home), strconv.FormatBool(houseEmpty))
	if err != nil {
		s.Logger.Error(err.Error())
		return err
	}
	body := "No longer Empty"
	if houseEmpty {
		body = fmt.Sprintf("No Humans in %s", home)
		devices, err := s.ReadNetworkConfig()
		if err != nil {
			s.Logger.Error(err.Error())
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
	}
	if !houseEmpty {
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
	return s.NotificationClient.SendNotification("House Empty", body, home)
}
