package config

import (
	"fmt"
	"os"

	"gopkg.in/ini.v1"
)

type SectionType string

const (
	APIDATA    SectionType = "api"
	HEADERDATA SectionType = "customHeader"
)

var (
	APIURL     string
	APIHEADERS Headers
)

type Config struct {
	Data interface{}
}

type Api struct {
	Url string `ini:"url"`
}

type Headers map[string]string

func init() {
	apiData, err := ReadConfig(APIDATA)
	if err != nil {
		fmt.Printf("Error getting config: %s\n", err)
		os.Exit(1)
	}
	apiInfo, ok := apiData.Data.(*Api)
	if !ok {
		fmt.Println("Error: could not cast Data to *Api")
		os.Exit(1)
	}
	if apiInfo.Url == "" {
		fmt.Println("Warning: API URL is empty!")
	}

	APIURL = apiInfo.Url

	headerData, err := ReadConfig(HEADERDATA)
	if err != nil {
		fmt.Printf("Error getting config: %s\n", err)
		os.Exit(1)
	}
	headerInfo, ok := headerData.Data.(*Headers)
	if !ok {
		fmt.Println("Error: could not cast Data to *Headers")
		os.Exit(1)
	}

	APIHEADERS = *headerInfo
}

func ReadConfig(section SectionType) (*Config, error) {
	c := new(Config)

	cfg, err := ini.Load("config/config.ini")
	if err != nil {
		return c, err
	}

	switch section {
	case APIDATA:
		api := new(Api)
		err = cfg.Section(string(section)).MapTo(api)
		c.Data = api
	case HEADERDATA:
		headers := Headers(cfg.Section(string(section)).KeysHash())
		c.Data = &headers
	default:
		return c, fmt.Errorf("Invalid section type: %s", section)
	}

	if err == nil {
		fmt.Printf("Mapped Struct: %+v\n", c)
	}

	return c, err
}

func SaveApi(url string) error {
	cfg, err := ini.Load("config/config.ini")
	if err != nil {
		return err
	}
	sec := cfg.Section(string(APIDATA))
	sec.Key("url").SetValue(url)
	if err := cfg.SaveTo("config/config.ini"); err != nil {
		return err
	}
	APIURL = url
	return nil
}

func SaveHeaders(headers map[string]string) error {
	cfg, err := ini.Load("config/config.ini")
	if err != nil {
		return err
	}
	cfg.DeleteSection(string(HEADERDATA))
	sec, err := cfg.NewSection(string(HEADERDATA))
	if err != nil {
		return err
	}
	for k, v := range headers {
		sec.Key(k).SetValue(v)
	}
	if err := cfg.SaveTo("config/config.ini"); err != nil {
		return err
	}
	APIHEADERS = headers
	return nil
}
