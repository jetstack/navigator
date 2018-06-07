package config

// config package provides an API with which Navigator components can read yaml
// or properties files and override certain keys before writing them back to the
// filesystem.
//
// This is currently a thin wrapper around ``viper`` which is already able to read and
// write files in multiple formats.
// The ``Interface`` has only the operations which are needed by Navigator, so
// we can swap viper something else if necessary in the future.
import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	shutil "github.com/termie/go-shutil"
)

const (
	BackupSuffix = ".navigator_original"
)

type config struct {
	*viper.Viper
}

type Interface interface {
	Get(key string) interface{}
	Set(key string, value interface{})
	Unset(key string)
	WriteConfig() error
	AllSettings() map[string]interface{}
}

var _ Interface = &config{}

func newFrom(path, format string) (Interface, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read absolute path")
	}
	c := &config{viper.New()}
	c.SetConfigFile(path)
	c.SetConfigType(format)
	f, err := os.Open(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, errors.Wrap(err, "unable to open file")
	}
	defer f.Close()
	err = c.ReadConfig(f)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read file")
	}
	return c, nil
}

// NewFromProperties reads a .yaml file.
func NewFromYaml(path string) (Interface, error) {
	return newFrom(path, "yaml")
}

// NewFromProperties reads a .properties file
func NewFromProperties(path string) (Interface, error) {
	return newFrom(path, "properties")
}

// Unset sets the value of the supplied ``key`` to ``null``
func (c *config) Unset(key string) {
	var null *struct{}
	c.Set(key, null)
}

// WriteConfig writes the configuration to file using the format of the original file.
// A backup is made, if the file already exists.
func (c *config) WriteConfig() error {
	path := c.ConfigFileUsed()
	backupPath := path + BackupSuffix
	err := shutil.CopyFile(path, backupPath, false)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "unable to create backup file")
	}
	return c.Viper.WriteConfig()
}
