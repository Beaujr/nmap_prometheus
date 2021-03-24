package house

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
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
	_, err = s.etcdClient.Put(context.Background(), key, string(d1))
	return err
}
