package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
)

var (
	metadataAddress = flag.String("metadata-address", "rancher-metadata", "The metadata service address")
)

func init() {
	logrus.SetOutput(os.Stdout)
}

func main() {
	flag.Parse()
	err := Poll(fmt.Sprintf("http://%s/2015-12-19", *metadataAddress))
	if err != nil {
		logrus.Fatal(err)
	}
	os.Exit(0)
}
