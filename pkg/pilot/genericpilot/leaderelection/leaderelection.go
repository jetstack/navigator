package leaderelection

import (
	"fmt"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
)

const (
	defaultLeaderElectionLeaseDuration = 15 * time.Second
	defaultLeaderElectionRenewDeadline = 10 * time.Second
	defaultLeaderElectionRetryPeriod   = 2 * time.Second
)

type Interface interface {
	Run() error
	Leading() bool
}

type Elector struct {
	LockMeta metav1.ObjectMeta
	Client   kubernetes.Interface
	Recorder record.EventRecorder

	leading bool
}

var _ Interface = &Elector{}

func (e *Elector) Run() error {
	// Identity used to distinguish between multiple controller manager instances
	id, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("error getting hostname: %s", err)
	}

	// Lock required for leader election
	rl := resourcelock.ConfigMapLock{
		ConfigMapMeta: e.LockMeta,
		Client:        e.Client.CoreV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity:      id + "-external-navigator-controller",
			EventRecorder: e.Recorder,
		},
	}

	leaderElector, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          &rl,
		LeaseDuration: defaultLeaderElectionLeaseDuration,
		RenewDeadline: defaultLeaderElectionRenewDeadline,
		RetryPeriod:   defaultLeaderElectionRetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(_ <-chan struct{}) { e.startLeading() },
			OnStoppedLeading: e.stopLeading,
		},
	})
	if err != nil {
		return err
	}
	// TODO: detect leader elector crashes
	leaderElector.Run()
	return nil
}

func (e *Elector) Leading() bool {
	return e.leading
}

func (e *Elector) startLeading() {
	e.leading = true
}

func (e *Elector) stopLeading() {
	e.leading = false
}
