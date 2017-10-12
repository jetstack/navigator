package v5

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/golang/glog"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
)

const (
	elasticsearchConfigSubDir = "elasticsearch/config"
)

func (p *Pilot) WriteConfig(pilot *v1alpha1.Pilot) error {
	esConfigPath := fmt.Sprintf("%s/%s", p.Options.ConfigDir, elasticsearchConfigSubDir)
	files, err := ioutil.ReadDir(esConfigPath)
	if err != nil {
		return fmt.Errorf("error listing provided config files: %s", err.Error())
	}
	for _, info := range files {
		path := filepath.Join(esConfigPath, info.Name())
		path, err = filepath.EvalSymlinks(path)
		if err != nil {
			return fmt.Errorf("error evaluating symlinks in path %q: %s", path, err.Error())
		}
		glog.V(2).Infof("Considering file %q (path: %q) when writing elasticsearch config", info.Name(), path)
		// re-check info after evaluating symlinks
		info, err = os.Stat(path)
		if err != nil {
			return fmt.Errorf("error getting info for path %q: %s", path, err.Error())
		}
		if info.IsDir() {
			continue
		}
		dstPath := fmt.Sprintf("%s/%s", p.Options.ElasticsearchOptions.ConfigDir, info.Name())
		glog.V(2).Infof("Copying config file from %q to %q", path, dstPath)
		if err = copyFileContents(path, dstPath); err != nil {
			return err
		}
	}
	if err != nil {
		return fmt.Errorf("error writing config file: %s", err.Error())
	}

	return nil
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
// From: https://stackoverflow.com/a/21067803
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
