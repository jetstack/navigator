package scheduler

import "time"

type FakeScheduledWorkQueue struct {
	AddFunc    func(interface{}, time.Duration)
	ForgetFunc func(interface{})
}

var _ ScheduledWorkQueue = &FakeScheduledWorkQueue{}

func (f *FakeScheduledWorkQueue) Add(i interface{}, t time.Duration) {
	f.AddFunc(i, t)
}

func (f *FakeScheduledWorkQueue) Forget(i interface{}) {
	f.ForgetFunc(i)
}
