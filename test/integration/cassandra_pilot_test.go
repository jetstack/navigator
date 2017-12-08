package cassandra_pilot_test

import (
	"testing"
	"time"

	"github.com/jetstack/navigator/test/integration/framework"
)

func TestCassandraPilot(t *testing.T) {
	t.Run(
		"cassandra pilot starts",
		func(t *testing.T) {
			f := framework.New(t)
			defer f.Close()
			masterUrl := f.RunAMaster()
			f.RunACassandraPilot(masterUrl)
			<-time.After(time.Second * 1)
		},
	)
}
