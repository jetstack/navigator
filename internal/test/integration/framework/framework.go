package framework

import (
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"time"

	"sigs.k8s.io/testing_frameworks/integration"

	"github.com/jetstack/navigator/internal/test/integration/framework/internal"
)

type NavigatorAPIServer struct {
	URL          *url.URL
	Path         string
	Args         []string
	StartTimeout time.Duration
	StopTimeout  time.Duration
	CertDir      string
	EtcdURL      *url.URL
	APIServerURL *url.URL
	Out          io.Writer
	Err          io.Writer
	processState *internal.ProcessState
}

func (s *NavigatorAPIServer) Start() error {
	if s.EtcdURL == nil {
		return fmt.Errorf("expected EtcdURL to be configured")
	}

	var err error

	s.processState = &internal.ProcessState{}

	s.processState.DefaultedProcessInput, err = internal.DoDefaulting(
		"navigator-apiserver",
		s.URL,
		s.CertDir,
		s.Path,
		s.StartTimeout,
		s.StopTimeout,
	)
	if err != nil {
		return err
	}

	// s.processState.HealthCheckEndpoint = "/healthz"
	s.processState.StartMessage = "Serving securely on 127.0.0.1"
	s.URL = &s.processState.URL
	s.CertDir = s.processState.Dir
	s.Path = s.processState.Path
	s.StartTimeout = s.processState.StartTimeout
	s.StopTimeout = s.processState.StopTimeout

	s.processState.Args, err = internal.RenderTemplates(
		append(
			internal.DoAPIServerArgDefaulting(nil),
			s.Args...,
		),
		s,
	)
	if err != nil {
		return err
	}
	return s.processState.Start(s.Out, s.Err)
}

func (s *NavigatorAPIServer) Stop() error {
	return s.processState.Stop()
}

type NavigatorControlPlane struct {
	*integration.ControlPlane
	NavigatorAPIServer *NavigatorAPIServer
}

func (f *NavigatorControlPlane) Start() error {
	if f.ControlPlane == nil {
		f.ControlPlane = &integration.ControlPlane{}
	}
	err := f.ControlPlane.Start()
	if err != nil {
		return err
	}
	if f.NavigatorAPIServer == nil {
		f.NavigatorAPIServer = &NavigatorAPIServer{}
	}
	f.NavigatorAPIServer.EtcdURL = f.Etcd.URL
	f.NavigatorAPIServer.APIServerURL = f.APIServer.URL
	return f.NavigatorAPIServer.Start()
}

func (f *NavigatorControlPlane) Stop() error {
	if f.NavigatorAPIServer != nil {
		err := f.NavigatorAPIServer.Stop()
		if err != nil {
			return err
		}
	}
	return f.ControlPlane.Stop()
}

func (f *NavigatorControlPlane) NavigatorAPIURL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   f.NavigatorAPIServer.URL.Host,
	}
}

func (f NavigatorControlPlane) NavigatorCtl() *NavigatorCtl {
	return &NavigatorCtl{
		&integration.KubeCtl{
			Opts: []string{
				"--server",
				f.NavigatorAPIURL().String(),
				"--client-certificate",
				filepath.Join(f.NavigatorAPIServer.CertDir, "apiserver.crt"),
				"--client-key",
				filepath.Join(f.NavigatorAPIServer.CertDir, "apiserver.key"),
				"--certificate-authority",
				filepath.Join(f.NavigatorAPIServer.CertDir, "apiserver.crt"),
			},
		},
	}
}

type NavigatorCtl struct {
	*integration.KubeCtl
}
