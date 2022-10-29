package helm

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type Helm struct {
	Config  Config
	Request Request
}

type Config struct {
}

type Request struct {
	RepoUrl string
	Body    url.Values
	Method  string
}

type HelmRepo struct {
	ApiVersion string                 `yaml:"apiVersion"`
	Entries    map[string][]HelmEntry `yaml:"entries"`
}

type HelmEntry struct {
	Created     time.Time
	Name        string
	Description string
	Digest      string
	Version     string
}

func (c *Helm) DoRequest(Req Request) (ResponseData []byte, StatusCode int, err error) {

	client := &http.Client{
		Timeout: time.Second * 5,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	apiurl := fmt.Sprintf("%s/index.yaml", c.Request.RepoUrl)
	_, err = url.ParseRequestURI(apiurl)
	if err != nil {
		return nil, 0, errors.New("Invalid Helm repo url")
	}

	req, _ := http.NewRequest(c.Request.Method, apiurl, strings.NewReader(c.Request.Body.Encode()))

	response, err := client.Do(req)
	if err != nil {
		log.Error().Str("Helm", "Response").Msg(err.Error())
		return nil, 0, err
	}
	defer response.Body.Close()
	responseData, _ := ioutil.ReadAll(response.Body)

	return responseData, response.StatusCode, nil
}

func (c *Helm) GetReleases(RepoUrl string) (HelmRepo, error) {

	c.Request = Request{
		RepoUrl: RepoUrl,
		Method:  "GET",
	}
	var ResponseObject HelmRepo

	data, code, err := c.DoRequest(c.Request)
	if code != http.StatusOK {
		return ResponseObject, err
	}

	yaml.Unmarshal(data, &ResponseObject)

	return ResponseObject, nil
}
