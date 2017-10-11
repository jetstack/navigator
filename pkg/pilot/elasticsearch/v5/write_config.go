package v5

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
)

const (
	elasticsearchConfigSubDir = "elasticsearch/config"
)

func (p *Pilot) WriteConfig(pilot *v1alpha1.Pilot) error {
	esConfigPath := fmt.Sprintf("%s/%s", p.Options.ConfigDir, elasticsearchConfigSubDir)
	err := filepath.Walk(esConfigPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), "..") {
				return filepath.SkipDir
			}
			return nil
		}
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(esConfigPath, path)
		if err != nil {
			return err
		}
		dstPath := fmt.Sprintf("%s/%s", p.Options.ElasticsearchOptions.ConfigDir, relPath)
		relDir := filepath.Dir(relPath)
		stat, err := os.Stat(relDir)
		if os.IsNotExist(err) {
			err = os.MkdirAll(relDir, 0644)
			if err != nil {
				return err
			}
			stat, err = os.Stat(relDir)
		}
		if err != nil {
			return err
		}
		if !stat.IsDir() {
			return fmt.Errorf("path '%s' exists and is not a directory", relDir)
		}
		if err = copyFileContents(path, dstPath); err != nil {
			return err
		}
		return nil
	})
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
