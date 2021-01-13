package etcd

import (
	"context"
	"fmt"
	etcdv3 "github.com/ozonru/etcd/v3/clientv3"
	"log"
	"time"
)

type Client struct {
	*etcdv3.Client
}

var (
	dialTimeout = 2 * time.Second
)

func NewClient(etcdServers []string) etcdv3.KV {
	cli, err := etcdv3.New(etcdv3.Config{
		DialTimeout: dialTimeout,
		Endpoints:   etcdServers,
	})
	//defer cli.Close()

	if err != nil {
		log.Println(err)
	}

	kv := etcdv3.NewKV(cli)
	//GetSingleValueDemo(context.Background(), kv)
	return kv
}

func GetSingleValueDemo(ctx context.Context, kv etcdv3.KV) {
	fmt.Println("*** GetSingleValueDemo()")
	// Delete all keys
	kv.Delete(ctx, "key", etcdv3.WithPrefix())

	// Insert a key value
	pr, _ := kv.Put(ctx, "key", "444")
	rev := pr.Header.Revision
	fmt.Println("Revision:", rev)

	gr, _ := kv.Get(ctx, "key")
	fmt.Println("Value: ", string(gr.Kvs[0].Value), "Revision: ", gr.Header.Revision)

	// Modify the value of an existing key (create new revision)
	kv.Put(ctx, "key", "555")

	gr, _ = kv.Get(ctx, "key")
	fmt.Println("Value: ", string(gr.Kvs[0].Value), "Revision: ", gr.Header.Revision)

	// Get the value of the previous revision
	gr, _ = kv.Get(ctx, "key", etcdv3.WithRev(rev))
	fmt.Println("Value: ", string(gr.Kvs[0].Value), "Revision: ", gr.Header.Revision)
}
