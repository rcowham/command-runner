// helpers/log.go

package helpers

import (
	"github.com/sirupsen/logrus"
)

func SetupLogger(debug bool) {
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}
