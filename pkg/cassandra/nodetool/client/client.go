package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pborman/uuid"

	"github.com/jetstack/navigator/pkg/cassandra/version"
)

const (
	storageServicePath = "read/org.apache.cassandra.db:type=StorageService"
)

type StorageService struct {
	HostIdMap        map[string]uuid.UUID `json:""`
	LiveNodes        []string             `json:""`
	UnreachableNodes []string             `json:""`
	LeavingNodes     []string             `json:""`
	JoiningNodes     []string             `json:""`
	MovingNodes      []string             `json:""`
	LocalHostId      uuid.UUID            `json:""`
	ReleaseVersion   *version.Version     `json:""`
}

type Interface interface {
	StorageService() (*StorageService, error)
}

type client struct {
	storageServiceURL *url.URL
	client            *http.Client
}

var _ Interface = &client{}

func New(baseURL *url.URL, c *http.Client) Interface {
	storageServiceURL, err := url.Parse(storageServicePath)
	if err != nil {
		panic(err)
	}
	storageServiceURL = baseURL.ResolveReference(storageServiceURL)
	return &client{
		storageServiceURL: storageServiceURL,
		client:            c,
	}
}

type JolokiaResponse struct {
	Value *StorageService `json:"value"`
}

func (c *client) StorageService() (*StorageService, error) {
	req, err := http.NewRequest(http.MethodGet, c.storageServiceURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate HTTP request: %v", err)
	}
	req.Header.Set("User-Agent", "navigator-cassandra-nodetool-client")

	response, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"unexpected server response code for request %s. "+
				"Expected %d. Got %d.",
			req.URL,
			http.StatusOK,
			response.StatusCode,
		)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	out := &JolokiaResponse{}

	err = json.Unmarshal(body, out)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON response: %v", err)
	}
	if out.Value == nil {
		return nil, fmt.Errorf("the response had an empty Jolokia value")
	}
	return out.Value, nil
}
