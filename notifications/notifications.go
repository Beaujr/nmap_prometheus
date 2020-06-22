package notifications

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var fcmUrl = flag.String("fcm", "http://fcmUrl", "Google Firbase Cloud Messaging URL")

func SendNotification(title string, message string) error {
	log.Printf("Notification: %s , %s", title, message)
	payload := strings.NewReader("{ \"title\": \"" + title + "\", \"body\":\"" + message + "\", \"image\": \"\"}")

	req, _ := http.NewRequest("POST", *fcmUrl, payload)
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
