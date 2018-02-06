package probe

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
)

type Listener struct {
	Port  int
	Check Check
}

// Listen will accept connections for this Probe
func (l *Listener) Listen() error {
	return http.ListenAndServe(
		fmt.Sprintf(":%d", l.Port),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := l.Check(); err != nil {
				glog.Errorf(
					"Error while running Check function for probe on port %d: %v",
					l.Port,
					err,
				)
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte(err.Error()))
				if err != nil {
					glog.Errorf(
						"Error while writing error message for probe on port %d: %v",
						l.Port,
						err,
					)
				}
			}
		}),
	)
}
