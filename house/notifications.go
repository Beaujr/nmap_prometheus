package house

import (
	"context"
	"flag"
	"fmt"
	etcdv3 "github.com/ozonru/etcd/v3/clientv3"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var fcmUrl = flag.String("fcm", "", "Google Firebase Cloud Messaging URL eg http://fcmUrl")

// Notifier represents and interface for notification sending
type Notifier interface {
	SendNotification(title string, message string, topic string) error
}

// NewNotifier returns a new Notifier
func NewNotifier() Notifier {
	if *debug || len(*fcmUrl) == 0 {
		return &DebugNotifier{}
	}
	return &FCMNotifier{url: fcmUrl}
}

// FCMNotifier is an implementation of the Notifier
type FCMNotifier struct {
	Notifier
	url *string
}

// DebugNotifier is a log implementation of the Notifier
type DebugNotifier struct {
	Notifier
}

func (s *Server) getLastSentNotification() (*string, error) {
	items, err := s.etcdClient.Get(context.Background(), fmt.Sprintf("%s%s", notificationsPrefix, "last"), etcdv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if items == nil {
		return nil, nil
	}

	if items.Count == 0 {
		return nil, nil
	}
	lastMessage := string(items.Kvs[0].Value)
	return &lastMessage, nil

}

func (s *Server) putLastSentNotification(notification string) error {
	_, err := s.etcdClient.Put(context.Background(), fmt.Sprintf("%s%s", notificationsPrefix, "last"), notification)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// SendNotification to GCM topic defined in fcmUrl
func (fcm *FCMNotifier) SendNotification(title string, message string, topic string) error {
	log.Printf("Notification: %s , %s", title, message)
	payload := strings.NewReader("{ \"title\": \"" + title + "\", \"body\":\"" + message + "\", \"image\": \"\"}")

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", *fcm.url, topic), payload)
	if err != nil {
		return err
	}
	req.Header.Add("content-type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return nil
}

// SendNotification to log output
func (fcm *DebugNotifier) SendNotification(title string, message string, topic string) error {
	log.Printf("Notification: %s , %s", title, message)
	return nil
}
