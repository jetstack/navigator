package processmanager

import "os"

type Signals struct {
	Stop      os.Signal
	Terminate os.Signal
	Reload    os.Signal
}
