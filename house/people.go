package house

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
)

type people struct {
	Id      string   `json:"id",yaml:"id"`
	Devices []string `json:"devices",yaml:"devices"`
	Name    string   `json:"name",yaml:"name"`
}

func (s *Server) writePerson(person *people) error {
	d1, err := yaml.Marshal(person)
	if err != nil {
		log.Fatalf(err.Error())
	}

	key := fmt.Sprintf("%s%s", peoplePrefix, person.Id)
	_, err = s.Kv.Put(context.Background(), key, string(d1))
	return err
}
