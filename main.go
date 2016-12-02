package main

import (
	"github.com/Sirupsen/logrus"
	"os"
)

func init() {
	logrus.SetOutput(os.Stdout)
}

func main() {
	err := Poll()
	if err != nil {
		logrus.Fatal(err)
	}
	os.Exit(0)
}
