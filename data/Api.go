package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/glgaspar/jui/config"
)

func Api(method, path string, data any) (*[]byte, error) {
	URL := config.APIURL
	HEADERS := config.APIHEADERS
	TOKEN := config.APITOKEN
	USER := config.APIUSER

	endpoint := URL + path

	var req *http.Request
	var err error

	if URL != "" && !strings.HasPrefix(URL, "http://") && !strings.HasPrefix(URL, "https://") {
		URL = "http://" + URL
	}

	if data != nil {
		payload := new(bytes.Buffer)
		if err = json.NewEncoder(payload).Encode(data); err != nil {
			return nil, err
		}
		req, err = http.NewRequest(method, endpoint, payload)
	} else {
		req, err = http.NewRequest(method, endpoint, nil)
	}
	if err != nil {
		return nil, err
	}

	for k, v := range HEADERS {
		req.Header.Add(k, v)
	}

	if USER != "" && TOKEN != "" {
		req.SetBasicAuth(USER, TOKEN)
	}

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return &body, fmt.Errorf("HTTP error: %s", response.Status)
	}
	return &body, nil
}
