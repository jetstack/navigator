package framework

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/jetstack/navigator/pkg/pilot/cassandra/v3"
	"k8s.io/apimachinery/pkg/util/wait"
)

type Framework struct {
	t       *testing.T
	closers []func()
}

func New(t *testing.T) *Framework {
	return &Framework{
		t: t,
	}
}

func (f *Framework) Close() {
	for _, closer := range f.closers {
		closer()
	}
}

func (f *Framework) addCloser(fn func()) {
	f.closers = append(f.closers, fn)
}

type CassandraNode struct {
}

func (f *Framework) RunAMaster() string {
	port := "8001"
	url := fmt.Sprintf("http://localhost:%s", port)
	cmd := exec.Command("kubectl", "proxy", "--port", port)
	err := cmd.Start()
	if err != nil {
		f.t.Fatal(err)
	}
	f.addCloser(
		func() {
			cmd.Process.Signal(syscall.SIGKILL)
			err := cmd.Wait()
			if err == nil {
				return
			}
			exiterr, ok := err.(*exec.ExitError)
			if ok {
				waitStatus := exiterr.Sys().(syscall.WaitStatus)
				if waitStatus.Signal() == syscall.SIGKILL {
					return
				}
			}
			f.t.Error(err)
		},
	)
	err = wait.Poll(
		time.Second*1,
		time.Second*5,
		func() (bool, error) {
			resp, err := http.Get(url)
			if err == nil {
				err = resp.Body.Close()
				if err != nil {
					f.t.Error(err)
				}
				return true, nil
			}
			return false, err
		},
	)
	if err != nil {
		f.t.Fatal(err)
	}
	return url
}

func (f *Framework) RunACassandraPilot(masterUrl string) {
	var wg sync.WaitGroup
	stopCh := make(chan struct{})
	f.addCloser(
		func() {
			close(stopCh)
			wg.Wait()
		},
	)
	cmd := v3.NewCommandStartPilot(os.Stdout, os.Stderr, stopCh)
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	cmd.ParseFlags(
		[]string{
			"--master", masterUrl,
			"--pilot-namespace", "default",
		},
	)
	flag.CommandLine.Parse([]string{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := cmd.Execute()
		if err != nil {
			f.t.Error(err)
		}
	}()
}
