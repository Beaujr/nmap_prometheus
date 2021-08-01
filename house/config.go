package house

import (
	"context"
	pb "github.com/beaujr/nmap_prometheus/proto"
	"github.com/ozonru/etcd/v3/clientv3"
	"gopkg.in/yaml.v2"
	"log"
)

func readServerConfig(etcdClient clientv3.KV) (*pb.ServerConfig, error) {
	items, err := etcdClient.Get(context.Background(), "/config")
	if err != nil {
		return nil, err
	}
	if items == nil {
		return nil, nil
	}

	if items.Count == 0 {
		err := writeServerConfig(&pb.ServerConfig{
			TimeAwaySeconds:    300,
			BleTimeAwaySeconds: 15,
			NewDeviceIsPerson:  false,
		}, etcdClient)
		if err == nil {
			return readServerConfig(etcdClient)
		}
		return nil, err
	}
	val := items.Kvs[0].Value
	var dev *pb.ServerConfig
	err = yaml.Unmarshal(val, &dev)
	if err != nil {
		return nil, err
	}
	return dev, nil
}

func writeServerConfig(item *pb.ServerConfig, etcdClient clientv3.KV) error {
	d1, err := yaml.Marshal(item)
	if err != nil {
		log.Fatalf(err.Error())
	}

	_, err = etcdClient.Put(context.Background(), "/config", string(d1))
	return err
}
