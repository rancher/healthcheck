package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/leodotcloud/log"
	logserver "github.com/leodotcloud/log/server"
)

var (
	metadataAddress = flag.String("metadata-address", "rancher-metadata", "The metadata service address")
)

func main() {
	logserver.StartServerWithDefaults()
	flag.Parse()
	err := Poll(fmt.Sprintf("http://%s/2015-12-19", *metadataAddress))
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
