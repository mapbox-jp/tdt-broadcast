package media

import (
	"gps_logger/logger"

	"github.com/aws/aws-sdk-go/service/mediastore"
)

func (m *Media) GetStore(mediaKey string) (string, error) {
	containerOutput, err := m.Ms.DescribeContainer(&mediastore.DescribeContainerInput{
		ContainerName: &mediaKey,
	})
	if err != nil {
		logger.Error("Failed to get store from redis. err: %v", err)
		return "", err
	}
	return *containerOutput.Container.Endpoint, nil
}
