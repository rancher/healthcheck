package util

import (
	"github.com/Sirupsen/logrus"
	rclient "github.com/rancher/go-rancher/client"
	"os"
)

func GetRancherClient() (*rclient.RancherClient, error) {
	cattleURL := os.Getenv("CATTLE_URL")
	if len(cattleURL) == 0 {
		logrus.Info("CATTLE_URL is not set")
		return nil, nil
	}

	cattleAccessKey := os.Getenv("CATTLE_ACCESS_KEY")
	if len(cattleAccessKey) == 0 {
		logrus.Info("CATTLE_ACCESS_KEY is not set")
		return nil, nil
	}

	cattleSecretKey := os.Getenv("CATTLE_SECRET_KEY")
	if len(cattleSecretKey) == 0 {
		logrus.Info("CATTLE_SECRET_KEY is not set")
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
