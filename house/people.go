package house

import (
	"context"
	"fmt"
	pb "github.com/beaujr/nmap_prometheus/proto"
	etcdv3 "github.com/ozonru/etcd/v3/clientv3"
	"gopkg.in/yaml.v2"
	"log"
)

var peoplePrefix = "/people/"

// Notifier represents and interface for notification sending
type PeopleManager interface {
	Get() ([]*pb.Person, error)
	CreateFromDevices(devices []*pb.Devices, name string) error
	Create(ids []string, name string) error
}

// NewNotifier returns a new Notifier
func NewPeopleManager(etcdClient etcdv3.KV) PeopleManager {
	return &EtcdPeopleManager{etcdClient: etcdClient}
}

// FCMNotifier is an implementation of the Notifier
type EtcdPeopleManager struct {
	Notifier
	etcdClient etcdv3.KV
}

func (etm *EtcdPeopleManager) Get() ([]*pb.Person, error) {
	items, err := etm.etcdClient.Get(context.Background(), fmt.Sprintf("%s", peoplePrefix), etcdv3.WithPrefix())
	people := make([]*pb.Person, 0)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return people, nil
	}

	for _, person := range items.Kvs {
		var dev *pb.Person
		err = yaml.Unmarshal(person.Value, &dev)
		if err != nil {
			return nil, err
		}
		people = append(people, dev)
	}
	return people, nil
}

func (etm *EtcdPeopleManager) CreateFromDevices(devices []*pb.Devices, name string) error {
	ids := make([]string, 0)
	for _, item := range devices {
		ids = append(ids, item.Id.Mac)
	}
	return etm.Create(ids, name)

}

func (etm *EtcdPeopleManager) Create(ids []string, name string) error {
	person := &pb.Person{
		Name: name,
		Ids:  ids,
	}
	return etm.writePerson(person)
}

func (etm *EtcdPeopleManager) writePerson(person *pb.Person) error {
	d1, err := yaml.Marshal(person)
	if err != nil {
		log.Fatalf(err.Error())
	}

	key := fmt.Sprintf("%s%s", peoplePrefix, person.Name)
	_, err = etm.etcdClient.Put(context.Background(), key, string(d1))
	return err
}
