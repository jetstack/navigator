package process

import "os"

type FakeAdapter struct {
	StartFunc     func() error
	StopFunc      func() error
	TerminateFunc func() error
	ReloadFunc    func() error
	WaitFunc      func() error
	RunningFunc   func() bool
	StateFunc     func() *os.ProcessState
	StringFunc    func() string
}

var _ Interface = &FakeAdapter{}

func (f *FakeAdapter) Start() error {
	return f.StartFunc()
}
func (f *FakeAdapter) Stop() error {
	return f.StopFunc()
}
func (f *FakeAdapter) Terminate() error {
	return f.TerminateFunc()
}
func (f *FakeAdapter) Reload() error {
	return f.ReloadFunc()
}
func (f *FakeAdapter) Wait() error {
	return f.WaitFunc()
}
func (f *FakeAdapter) Running() bool {
	return f.RunningFunc()
}
func (f *FakeAdapter) State() *os.ProcessState {
	return f.StateFunc()
}
func (f *FakeAdapter) String() string {
	return f.StringFunc()
}
