package house

import (
	"context"
	"flag"
	"fmt"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var fcmUrl = flag.String("fcm", "", "Google Firbase Cloud Messaging URL eg http://fcmUrl")

// Notifier represents and interface for notification sending
type Notifier interface {
	SendNotification(title string, message string, topic string) error
}

// NewNotifier returns a new Notifier
func NewNotifier(etcdClient etcdv3.KV) Notifier {
	if *debug || len(*fcmUrl) == 0 {
		return &DebugNotifier{}
	}
	return &FCMNotifier{url: fcmUrl, etcdClient: etcdClient}
}

// FCMNotifier is an implementation of the Notifier
type FCMNotifier struct {
	Notifier
	etcdClient etcdv3.KV
	url        *string
}

// DebugNotifier is a log implementation of the Notifier
type DebugNotifier struct {
	Notifier
}

func (fcm *FCMNotifier) getLastSentNotification() (*string, error) {
	items, err := fcm.etcdClient.Get(context.Background(), fmt.Sprintf("%s%s", notificationsPrefix, "last"), etcdv3.WithPrefix())
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

func (fcm *FCMNotifier) putLastSentNotification(notification string) error {
	_, err := fcm.etcdClient.Put(context.Background(), fmt.Sprintf("%s%s", notificationsPrefix, "last"), notification)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// SendNotification to GCM topic defined in fcmUrl
func (fcm *FCMNotifier) SendNotification(title string, message string, topic string) error {
	lastNotification, err := fcm.getLastSentNotification()
	currentNotificationKey := fmt.Sprintf("%s%s%s", title, message, topic)

	if lastNotification != nil &&
		strings.Compare(*lastNotification, currentNotificationKey) == 0 {
		return nil
	}

	log.Printf("Notification: %s , %s", title, message)
	payload := strings.NewReader("{ \"title\": \"" + title + "\", \"body\":\"" + message + "\", \"image\": \"\"}")

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", *fcm.url, topic), payload)
	if err != nil {
		log.Println(err.Error())
	}
	req.Header.Add("content-type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err.Error())
	}

	defer res.Body.Close()
	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err.Error())
	}

	if err = fcm.putLastSentNotification(currentNotificationKey); err != nil {
		log.Printf("failed saving notification: %s with error: %s", currentNotificationKey, err.Error())
	}
	return nil
}

// SendNotification to log output
func (fcm *DebugNotifier) SendNotification(title string, message string, topic string) error {
	log.Printf("Notification: %s , %s", title, message)
	return nil
}
