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

//func (s *Server) readHomesConfig() (map[string]*home, error) {
//	var result map[string]*home
//	result = make(map[string]*home)
//	items, err := s.etcdClient.Get(context.Background(), homePrefix, clientv3.WithPrefix())
//	if err != nil {
//		return nil, err
//	}
//	if items == nil {
//		return result, nil
//	}
//	i := 0
//	for i < int(items.Count) {
//		val := items.Kvs[i].Value
//		key := items.Kvs[i].Key
//		var home *home
//		err = yaml.Unmarshal(val, &home)
//		if err != nil {
//			return nil, err
//		}
//		newKey := strings.ReplaceAll(string(key), homePrefix, "")
//		//boolVal, _ := strconv.ParseBool(string(val))
//		if strings.Contains(string(key), "//") {
//			key2 := strings.ReplaceAll(string(key), "//", "")
//			s.etcdClient.Put(context.Background(), key2, string(val))
//			s.etcdClient.Delete(context.Background(), string(key))
//		}
//		result[string(newKey)] = home
//		i++
//	}
//	return result, nil
//}

func (s *Server) writeHome(id string, item *home) error {
	d1, err := yaml.Marshal(item)
	if err != nil {
		log.Fatalf(err.Error())
	}

	key := fmt.Sprintf("%s%s", homePrefix, id)
	_, err = s.etcdClient.Put(context.Background(), key, string(d1))
	return err
}
