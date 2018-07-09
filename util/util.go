package util

import (
	"os"

	rclient "github.com/rancher/go-rancher/client"
	"github.com/rancher/log"
)

func GetRancherClient() (*rclient.RancherClient, error) {
	cattleURL := os.Getenv("CATTLE_URL")
	if len(cattleURL) == 0 {
		log.Info("CATTLE_URL is not set")
		return nil, nil
	}

	cattleAccessKey := os.Getenv("CATTLE_ACCESS_KEY")
	if len(cattleAccessKey) == 0 {
		log.Info("CATTLE_ACCESS_KEY is not set")
		return nil, nil
	}

	cattleSecretKey := os.Getenv("CATTLE_SECRET_KEY")
	if len(cattleSecretKey) == 0 {
		log.Info("CATTLE_SECRET_KEY is not set")
		return nil, nil
	}

	apiClient, err := rclient.NewRancherClient(&rclient.ClientOpts{
		Url:       cattleURL,
		AccessKey: cattleAccessKey,
		SecretKey: cattleSecretKey,
	})
	if err != nil {
		return nil, err
	}
	return apiClient, nil
}
