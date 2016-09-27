package main

import (
	"github.com/Sirupsen/logrus"
	"os"
)

func main() {
	err := Poll()
	if err != nil {
		logrus.Fatal(err)
	}
	os.Exit(0)
}
