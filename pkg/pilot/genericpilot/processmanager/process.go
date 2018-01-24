package processmanager

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/golang/glog"
)

type Interface interface {
	// Start will start the underlying subprocess
	Start() error
	// Stop will stop the underlying subprocess gracefully and wait for it
	// to exit
	Stop() error
	// Wait will return a channel that closes upon the subprocess exiting.
	// If the subprocess has not been started yet, the returned chan will
	// not close until the subprocess has been started and then stopped.
	Wait() <-chan struct{}
	// Running returns true if the subprocess is currently running
	Running() bool
	// String returns a string representation of this process
	String() string
	// Error returns the error returned from Wait() (or nil if there was no
	// error). If the process has not been started or is still running, nil
	// will be returned.
	Error() error
}

func New(cmd *exec.Cmd, signals Signals) Interface {
	return &adapter{
		cmd:     cmd,
		signals: signals,
		doneCh:  make(chan struct{}),
	}
}

// Adapter makes it easy to create managed subprocesses
type adapter struct {
	signals Signals
	cmd     *exec.Cmd

	doneCh  chan struct{}
	doneErr error
	wg      sync.WaitGroup
}

var _ Interface = &adapter{}

func (p *adapter) startCommandOutputLoggers() error {
	stdout, err := p.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		in := bufio.NewScanner(stdout)
		for in.Scan() {
			glog.Infoln(in.Text())
		}
		err := in.Err()
		if err != nil {
			glog.Error(err)
		}
	}()

	stderr, err := p.cmd.StderrPipe()
	if err != nil {
		return err
	}
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		in := bufio.NewScanner(stderr)
		for in.Scan() {
			glog.Errorln(in.Text())
		}
		err := in.Err()
		if err != nil {
			glog.Error(err)
		}
	}()
	return nil
}

// Start will start the underlying subprocess
func (p *adapter) Start() error {
	err := p.startCommandOutputLoggers()
	if err != nil {
		return err
	}

	if err = p.cmd.Start(); err != nil {
		return fmt.Errorf("error starting process: %s", err.Error())
	}
	go p.startWait()
	return nil
}

// Stop will stop the underlying subprocess gracefully and wait for it to
// exit. It returns the error returned from calling Wait(), or no error if the
// process has not been started.
func (p *adapter) Stop() error {
	if p.Running() {
		if err := p.cmd.Process.Signal(p.signals.Stop); err != nil {
			return err
		}
		<-p.Wait()
	}
	return p.Error()
}

// Wait will return a channel that closes upon the subprocess exiting.
// If the subprocess has not been started yet, the returned chan will
// not close until the subprocess has been started and then stopped.
func (p *adapter) Wait() <-chan struct{} {
	defer p.wg.Wait()
	return p.doneCh
}

// Running returns true if the subprocess is currently running
func (p *adapter) Running() bool {
	if p.cmd == nil || p.cmd.Process == nil || p.cmd.Process.Pid == 0 || p.state() != nil {
		return false
	}
	return true
}

// String returns a string representation of this process
func (p *adapter) String() string {
	if !p.Running() {
		return fmt.Sprintf("inactive")
	}
	return fmt.Sprintf("%d", p.cmd.Process.Pid)
}

func (p *adapter) Error() error {
	return p.doneErr
}

func (p *adapter) state() *os.ProcessState {
	if p.cmd == nil {
		return nil
	}
	return p.cmd.ProcessState
}

func (p *adapter) startWait() {
	p.doneErr = p.cmd.Wait()
	close(p.doneCh)
}
