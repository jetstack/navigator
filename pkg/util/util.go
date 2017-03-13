package util

import "github.com/Sirupsen/logrus"

// HandleError will gracefully handle a fatal error within a control loop and exit
func HandleError(err error) {
	logrus.Fatalf("error running colonel: %s", err.Error())
}
