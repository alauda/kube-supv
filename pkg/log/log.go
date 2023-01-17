package log

import "github.com/sirupsen/logrus"

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{})
}

func Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}
