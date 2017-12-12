package framework

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/golang/glog"
	"github.com/jetstack/navigator/pkg/pilot/elasticsearch/v5"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/wait"
)

type Framework struct {
	t       *testing.T
	closers []<-chan struct{}
}

func New(t *testing.T) *Framework {
	return &Framework{
		t: t,
	}
}

func (f *Framework) Wait() {
	for _, c := range f.closers {
		<-c
	}
}

func (f *Framework) addCloser(c chan struct{}) {
	f.closers = append(f.closers, c)
}

func http_responding(url string) (bool, error) {
	resp, err := http.Get(url)
	if err != nil {
		glog.Error(err)
		return false, nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return true, err
	}
	glog.Infof("HTTP response from %s, %s, %s", url, resp, body)
	err = resp.Body.Close()
	return true, err
}

// RunAMaster would ideally start an in-memory K8S API server, but for now it
// starts a proxy to the master for the current kubectl context.
func (f *Framework) RunAMaster(ctx context.Context) string {
	ctx, cancel := context.WithCancel(ctx)
	// TODO: Pick a free local port instead.
	port := "8001"
	url := fmt.Sprintf("http://localhost:%s", port)
	cmd := exec.CommandContext(ctx, "kubectl", "proxy", "--port", port)
	glog.Infof("Starting kubectl proxy (%v)", cmd)
	err := cmd.Start()
	if err != nil {
		f.t.Fatal(err)
	}
	finished := make(chan struct{})
	f.addCloser(finished)
	go func() {
		defer close(finished)
		err := cmd.Wait()
		glog.Infof("kubectl proxy exited with. %v", err)
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
	}()
	err = wait.Poll(
		time.Second*1,
		time.Second*10,
		func() (bool, error) {
			glog.Infof("Waiting for kubectl proxy port to accept connections. %v", url)
			select {
			case <-finished:
				return false, fmt.Errorf("kubectl proxy exited without accepting connections")
			default:
				return http_responding(url)
			}
		},
	)
	if err != nil {
		cancel()
		f.t.Fatal(err)
	}
	return url
}

func (f *Framework) RunACobraCommand(ctx context.Context, cmd *cobra.Command) chan struct{} {
	stopCh := make(chan struct{})
	f.addCloser(stopCh)
	err := flag.CommandLine.Parse([]string{})
	if err != nil {
		f.t.Fatal(err)
	}
	go func() {
		defer glog.Infof("Cobra command finished. %v", cmd)
		defer close(stopCh)
		glog.Infof("Executing cobra command. %v", cmd)
		err := cmd.Execute()
		if err != nil {
			f.t.Error(err)
		}
	}()
	return stopCh
}

func (f *Framework) RunAnElasticSearchPilot(ctx context.Context, masterUrl string) chan struct{} {
	ctx, cancel := context.WithCancel(ctx)
	cmd := v5.NewCommandStartPilot(os.Stdout, os.Stderr, ctx.Done())
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	cmd.ParseFlags(
		[]string{
			"--master", masterUrl,
			"--pilot-namespace", "default",
			"--elasticsearch-master-url", "http://192.0.2.100",
		},
	)
	stopCh := f.RunACobraCommand(ctx, cmd)
	err := wait.Poll(
		time.Second*1,
		time.Second*10,
		func() (bool, error) {
			glog.Infof("Waiting for ES pilot to accept healthz requests.")
			select {
			case <-stopCh:
				return false, fmt.Errorf("pilot exited before accept connections")
			default:
				return http_responding("http://localhost:12000")
			}

		},
	)
	if err != nil {
		cancel()
		f.t.Fatal(err)
	}
	return stopCh
}
