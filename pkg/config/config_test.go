package config_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	shutil "github.com/termie/go-shutil"

	"github.com/jetstack/navigator/internal/test/util/testfs"
	"github.com/jetstack/navigator/pkg/config"
)

// PathExists returns true if the path exists (file, directory or symlink)
func PathExists(path string) (bool, error) {
	_, err := os.Lstat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// AssertNotPathExists asserts that a path does not exist.
// This is missing from the testify library.
func AssertNotPathExists(t *testing.T, path string) {
	exists, err := PathExists(path)
	require.NoError(t, err)
	if exists {
		assert.Fail(t, "AssertNotPathExists", fmt.Sprintf("Path %q exists", path))
	}
}

// TestUnset verifies that Unset sets the supplied key to null / nil
func TestUnset(t *testing.T) {
	tfs := testfs.New(t)
	inPath := tfs.TempPath("config.yaml")
	err := shutil.CopyFile("testdata/config1.yaml", inPath, false)
	require.NoError(t, err)
	c1, err := config.NewFromYaml(inPath)
	require.NoError(t, err)
	assert.NotNil(t, "a.b.c")
	c1.Unset("a.b.c")
	err = c1.WriteConfig()
	require.NoError(t, err)
	c2, err := config.NewFromYaml(inPath)
	require.NoError(t, err)
	assert.Nil(t, c2.Get("a.b.c"))
}

// The file written by WriteConfig can be re-read and results in identical settings.
func TestRoundTrip(t *testing.T) {
	tfs := testfs.New(t)

	inPath := tfs.TempPath("config.yaml")
	err := shutil.CopyFile("testdata/config1.yaml", inPath, false)
	require.NoError(t, err)

	c1, err := config.NewFromYaml(inPath)
	require.NoError(t, err)

	err = c1.WriteConfig()
	require.NoError(t, err)

	c2, err := config.NewFromYaml(inPath)
	require.NoError(t, err)

	assert.Equal(t, c1.AllSettings(), c2.AllSettings())
}

// TestBackup verifies that backup files are created when an existing
// configuration file already exists.
func TestBackup(t *testing.T) {
	t.Run(
		"Backup file is created only when the config is written",
		func(t *testing.T) {
			tfs := testfs.New(t)
			path := tfs.TempPath("conf1.yaml")
			cfg, err := config.NewFromYaml(path)
			require.NoError(t, err)
			backupPath := path + config.BackupSuffix
			AssertNotPathExists(t, backupPath)
			err = cfg.WriteConfig()
			require.NoError(t, err)
			assert.FileExists(t, backupPath)
		},
	)
	t.Run(
		"Backup file is only created if a config file already exists",
		func(t *testing.T) {
			tfs := testfs.New(t)
			dir := tfs.TempDir("conf")
			// A non-existent filepath
			path := dir + "/conf1.yaml"
			cfg, err := config.NewFromYaml(path)
			require.NoError(t, err)
			backupPath := path + config.BackupSuffix
			AssertNotPathExists(t, backupPath)
			err = cfg.WriteConfig()
			require.NoError(t, err)
			AssertNotPathExists(t, backupPath)
		},
	)
}
