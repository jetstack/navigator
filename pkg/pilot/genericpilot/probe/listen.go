package probe

import (
	"fmt"
	"net/http"
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
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			}
		}),
	)
}
