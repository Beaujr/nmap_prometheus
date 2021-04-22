package house

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	assistantPath          = "assistant"
	pocketSphinxUrl        = "http://localhost:8085/stt"
	tvIsOnLength           = 64512
	deskIsOnLength         = 46840
	boxRoomLightIsOnLength = 56878
	livingLightIsOnLength  = 322078
	tvCheck                = "IS THE TV ON?"
	deskCheck              = "IS THE DESK ON?"
	boxroomLightCheck      = "IS LIGHT BULB ON?"
	loungeLightCheck       = "Are the Living room lights on?"
)

var (
	assistantUrl  = flag.String("assistant", "", "Google Assistant URL eg http://assistant_relay")
	assistantUser = flag.String("assistantUser", "", "Google Assistant Relay User")
)

type gaResponse struct {
	Response string `json:"response"`
	Audio    string `json:"audio"`
	Success  bool   `json:"success"`
}

type psResponse struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

// GoogleAssistant interface for calling smart home api
type GoogleAssistant interface {
	Call(command string) (*string, error)
}

// NewAssistant returns a new assistant client
func NewAssistant() GoogleAssistant {
	if *debug || len(*assistantUser) == 0 || len(*assistantUrl) == 0 {
		return &DebugAssistantRelay{}
	}
	return &AssistantRelay{url: assistantUrl, path: assistantPath, user: assistantUser}
}

// AssistantRelay is an implementation of the GoogleAssistant
type AssistantRelay struct {
	GoogleAssistant
	url  *string
	path string
	user *string
}

// DebugAssistantRelay is an implementation of the GoogleAssistant which only logs
type DebugAssistantRelay struct {
	GoogleAssistant
}

// Call assistant relay with command
func (ga *AssistantRelay) Call(command string) (*string, error) {
	payload := strings.NewReader("{\"user\":\"" + *ga.user + "\",\"command\":\"" + command + "\", \"converse\": false}")
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", *ga.url, assistantPath), payload)
	if err != nil {
		log.Printf("Error: %s\n", err)
		return nil, err
	}
	req.Header.Add("content-type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error: %s\n", err)
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("error executing %s", command)
	}
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)
	assistantResponse := gaResponse{}
	if err := json.Unmarshal(body, &assistantResponse); err != nil {
		panic(err)
	}

	if assistantResponse.Response == "" {
		ps, err := downloadFile(assistantResponse.Audio, command)
		if err != nil {
			return nil, err
		}
		tts := ps.Text
		return &tts, nil
	}
	return &assistantResponse.Response, nil
}

// Call to log the command
func (ga *DebugAssistantRelay) Call(command string) (*string, error) {
	log.Println(command)
	return &command, nil
}

func downloadFile(url string, command string) (*psResponse, error) {

	// Get the data
	fcmServerUrl := fmt.Sprintf("%s%s", *assistantUrl, url)
	log.Println(fcmServerUrl)
	resp, err := http.Get(fcmServerUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if command == tvCheck {
		if resp.ContentLength == tvIsOnLength {
			return &psResponse{Text: "on"}, err
		}
		return &psResponse{Text: "off"}, err
	}
	if command == deskCheck {
		if resp.ContentLength == deskIsOnLength {
			return &psResponse{Text: "on"}, err
		}
		return &psResponse{Text: "off"}, err
	}
	if command == boxroomLightCheck {
		if resp.ContentLength == boxRoomLightIsOnLength {
			return &psResponse{Text: "on"}, err
		}
		return &psResponse{Text: "off"}, err
	}
	if command == loungeLightCheck {
		if resp.ContentLength == livingLightIsOnLength {
			return &psResponse{Text: "on"}, err
		}
		return &psResponse{Text: "off"}, err
	}

	return &psResponse{Text: "success"}, nil
}

func callPocketSphinx(body io.ReadCloser) (*psResponse, error) {
	request, err := http.NewRequest("POST", pocketSphinxUrl, body)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	request.Header.Add("Content-Type", "audio/x-wav")
	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	assistantResponse := psResponse{}
	if err := json.Unmarshal(content, &assistantResponse); err != nil {
		panic(err)
	}
	return &assistantResponse, nil
}
