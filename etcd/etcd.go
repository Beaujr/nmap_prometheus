package etcd

import (
	etcdv3 "go.etcd.io/etcd/client/v3"
	"log"
	"time"
)

// Client is a wrapped etcdv3.Client
type Client struct {
	*etcdv3.Client
}

var (
	dialTimeout = 2 * time.Second
)

// NewClient returns a new etcdv3.Client
func NewClient(etcdServers []string) (*etcdv3.Client, etcdv3.KV) {
	cli, err := etcdv3.New(etcdv3.Config{
		DialTimeout: dialTimeout,
		Endpoints:   etcdServers,
	})
	//defer cli.Close()

	if err != nil {
		log.Println(err)
	}

	kv := etcdv3.NewKV(cli)
	return cli, kv
}
