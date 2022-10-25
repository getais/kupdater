package helm

import (
	"crypto/tls"
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

func (c *Helm) DoRequest(Req Request) (ResponseData []byte, StatusCode int) {

	client := &http.Client{
		Timeout: time.Second * 5,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	apiurl := fmt.Sprintf("%s/index.yaml", c.Request.RepoUrl)

	req, _ := http.NewRequest(c.Request.Method, apiurl, strings.NewReader(c.Request.Body.Encode()))
	// req.Header.Set("Accept", "application/json")

	response, err := client.Do(req)
	if err != nil {
		log.Error().Str("Helm", "Response").Msg(err.Error())
	}
	defer response.Body.Close()
	responseData, _ := ioutil.ReadAll(response.Body)

	return responseData, response.StatusCode
}

func (c *Helm) GetReleases(RepoUrl string) HelmRepo {

	c.Request = Request{
		RepoUrl: RepoUrl,
		Method:  "GET",
	}

	data, _ := c.DoRequest(c.Request)

	// var ResponseObject struct {
	// 	Release
	// }
	var ResponseObject HelmRepo
	yaml.Unmarshal(data, &ResponseObject)

	return ResponseObject
}
