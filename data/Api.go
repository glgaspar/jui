package data

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/glgaspar/jui/config"
)

func Api(method, path string, data any) (*[]byte, error) {
	URL := config.APIURL
	HEADERS := config.APIHEADERS

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
		req, err = http.NewRequest(method, URL+path, payload)
	} else {
		req, err = http.NewRequest(method, URL+path, nil)
	}
	if err != nil {
		return nil, err
	}

	for k, v := range HEADERS {
		req.Header.Add(k, v)
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
		return &[]byte{}, err
	}
	return &body, nil
}
