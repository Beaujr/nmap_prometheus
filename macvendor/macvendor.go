package macvendor

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// GetManufacturer to get the vendor of the device
func GetManufacturer(mac string) (*string, error) {
	//log.Printf("Notification: %s , %s", title, message)
	//payload := strings.NewReader("{ \"title\": \"" + title + "\", \"body\":\"" + message + "\", \"image\": \"\"}")
	requestURL := fmt.Sprintf("https://api.macvendors.com/%s", mac)
	req, _ := http.NewRequest("GET", requestURL, nil)
	//req.Header.Add("content-type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode == 429 {
		return nil, fmt.Errorf("MacVendors time out: %s", mac)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	vendor := string(body)
	return &vendor, nil
}
