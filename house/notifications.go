package house

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var fcmUrl = flag.String("fcm", "", "Google Firbase Cloud Messaging URL eg http://fcmUrl")

// Notifier
type Notifier interface {
	SendNotification(title string, message string, topic string) error
}

func NewNotifier() Notifier {
	if *debug || len(*fcmUrl) == 0 {
		return &DebugNotifier{}
	}
	return &FCMNotifier{url: fcmUrl}
}

// Server is an implementation of the proto HomeDetectorServer
type FCMNotifier struct {
	Notifier
	url *string
}

// Server is an implementation of the proto HomeDetectorServer
type DebugNotifier struct {
	Notifier
}

// SendNotification to GCM topic defined in fcmUrl
func (fcm *FCMNotifier) SendNotification(title string, message string, topic string) error {
	log.Printf("Notification: %s , %s", title, message)
	payload := strings.NewReader("{ \"title\": \"" + title + "\", \"body\":\"" + message + "\", \"image\": \"\"}")

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s%s", *fcm.url, topic), payload)
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

func (fcm *DebugNotifier) SendNotification(title string, message string, topic string) error {
	log.Printf("Notification: %s , %s", title, message)
	return nil
}
