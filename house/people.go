package house

import (
	"context"
	"fmt"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v2"
	"log"
	"path/filepath"
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
	_, err = s.Kv.Put(s.GetContext(), key, string(d1))
	return err
}

func (s *Server) GetPeopleInHouses(ctx context.Context, home string) ([]string, error) {
	var people []string
	items, err := s.Kv.Get(ctx, filepath.Join(AlivePrefix, home), etcdv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if items == nil {
		return people, nil
	}
	i := 0
	for i < int(items.Count) {
		val := items.Kvs[i].Value
		key := items.Kvs[i].Key
		id := filepath.Base(string(key))
		if string(val) == "person" {
			people = append(people, id)
		}
		i++
	}
	return people, nil
}
