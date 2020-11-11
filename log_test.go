package mylogger

import (
	"github.com/sirupsen/logrus"
	"os"
	"testing"
)

func init() {
	logrus.SetFormatter(&MyJSONFormatter{})
	logrus.SetFormatter(&MyTextFormatter{})
	logrus.SetOutput(os.Stdout)
}

func TestNewConsumer(t *testing.T) {
	file, err := os.OpenFile("./logger.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 7777)
	if err == nil {
		logrus.SetOutput(file)
	} else {
		logrus.Fatal("Failed to log to file, using default stderr", err)
	}

	logrus.Info("test ok")

	t.Fail()
}
