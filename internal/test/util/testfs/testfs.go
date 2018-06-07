package testfs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestFs struct {
	t *testing.T
	d string
}

func New(t *testing.T) *TestFs {
	d := fmt.Sprintf(".test/%s", t.Name())
	d, err := filepath.Abs(d)
	require.NoError(t, err)
	err = os.RemoveAll(d)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("Error while removing old test directory: %s", err)
	}

	err = os.MkdirAll(d, os.ModePerm)
	require.NoError(t, err)

	return &TestFs{
		t: t,
		d: d,
	}
}

func (tfs *TestFs) TempPath(name string) string {
	outPath := path.Join(tfs.d, name)
	tmpFile, err := ioutil.TempFile(tfs.d, name)
	require.NoError(tfs.t, err)
	err = tmpFile.Close()
	require.NoError(tfs.t, err)
	err = os.Rename(tmpFile.Name(), outPath)
	require.NoError(tfs.t, err)
	return outPath
}

func (tfs *TestFs) TempDir(name string) string {
	outPath := path.Join(tfs.d, name)
	err := os.MkdirAll(outPath, os.ModePerm)
	require.NoError(tfs.t, err)
	return outPath
}
