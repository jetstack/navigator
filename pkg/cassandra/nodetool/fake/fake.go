package fake

import (
	"fmt"

	"github.com/jetstack/navigator/pkg/cassandra/nodetool"
	"github.com/jetstack/navigator/pkg/cassandra/nodetool/client"
	"github.com/jetstack/navigator/pkg/cassandra/version"
)

type FakeNodeTool struct {
	StatusResult  nodetool.NodeMap
	StatusError   error
	VersionResult *version.Version
	VersionError  error
}

var _ nodetool.Interface = &FakeNodeTool{}

func New() *FakeNodeTool {
	return &FakeNodeTool{}
}

func (nt *FakeNodeTool) Status() (nodetool.NodeMap, error) {
	return nt.StatusResult, nt.StatusError
}

func (nt *FakeNodeTool) Version() (*version.Version, error) {
	return nt.VersionResult, nt.VersionError
}

func (nt *FakeNodeTool) SetVersion(v string) *FakeNodeTool {
	nt.VersionResult = version.New(v)
	return nt
}

func (nt *FakeNodeTool) SetVersionError(e string) *FakeNodeTool {
	nt.VersionError = fmt.Errorf(e)
	return nt
}

type FakeClient struct {
	StorageServiceResult *client.StorageService
	StorageServiceError  error
}

var _ client.Interface = &FakeClient{}

func NewClient() *FakeClient {
	return &FakeClient{}
}

func (c *FakeClient) StorageService() (*client.StorageService, error) {
	return c.StorageServiceResult, c.StorageServiceError
}

func (c *FakeClient) SetReleaseVersion(v string) *FakeClient {
	if c.StorageServiceResult == nil {
		c.StorageServiceResult = &client.StorageService{}
	}
	c.StorageServiceResult.ReleaseVersion = version.New(v)
	return c
}

func (c *FakeClient) SetStorageServiceError(e string) *FakeClient {
	c.StorageServiceError = fmt.Errorf(e)
	return c
}
