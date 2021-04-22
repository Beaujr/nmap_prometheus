package house

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
)

// GetManufacturer to get the vendor of the device
func GetManufacturer(mac string) (*string, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	requestURL := fmt.Sprintf("https://api.macvendors.com/%s", mac)
	req, _ := http.NewRequest("GET", requestURL, nil)
	res, err := client.Do(req)
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
