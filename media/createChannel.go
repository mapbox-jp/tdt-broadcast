package media

import (
	"encoding/json"
	"fmt"
	"gps_logger/logger"
	"io/ioutil"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/service/medialive"
	"github.com/aws/aws-sdk-go/service/mediastore"
)

func (m *Media) CreateChannel() (string, error) {
	var err error
	ml := medialive.New(m.Sess)
	ms := mediastore.New(m.Sess)
	time := time.Now()
	mediaKey := fmt.Sprintf("%d%02d%02d%02d%02d", time.Year(), time.Month(), time.Day(), time.Hour(), time.Minute())

	// --------- Create Media Store Container --------- //
	logger.Info("Started creating a media container. key: %v, endpoint: %v", mediaKey)
	// Create Params
	containerInput := &mediastore.CreateContainerInput{
		ContainerName: &mediaKey,
	}
	// Create Container
	_, err = ms.CreateContainer(containerInput)
	if err != nil {
		logger.Error("Failed creating the store container. key: %v, err: %v", mediaKey, err)
		return "", err
	}
	// Get Store Container Result
	describeContainerOutput, err := m.describeContainer(mediaKey)
	if err != nil {
		logger.Error("Failed creating the store container. key: %v, err: %v", mediaKey, err)
		return "", err
	}
	logger.Info("Suceeded creating a media container. key: %v", mediaKey)

	// --------- Create Media Live Input --------- //
	logger.Info("Started creating a media input. key: %v, endpoint: %v", mediaKey)
	// Create Params
	input_template_raw, err := ioutil.ReadFile("./input_template.json")
	if err != nil {
		logger.Error("Failed reading the template json to create a media input. key: %v, err: %v", mediaKey, err)
		return "", err
	}
	var inputParams *medialive.CreateInputInput
	if err := json.Unmarshal(input_template_raw, &inputParams); err != nil {
		logger.Error("Failed creating a media input. key: %v, err: %v", mediaKey, err)
		return "", err
	}
	inputParams.Name = &mediaKey
	// Create Channel Input
	inputOutput, err := ml.CreateInput(inputParams)
	if err != nil {
		logger.Error("Failed creating channel key: %v, err: %v", mediaKey, err)
		return "", err
	}
	logger.Info("Suceeded creating a media input. key: %v", mediaKey)

	// --------- Create Media Live Channel --------- //
	logger.Info("Started creating a media channel. key: %v, endpoint: %v", mediaKey)
	// Create Params
	channel_template_raw, err := ioutil.ReadFile("./channel_template.json")
	if err != nil {
		return "", err
	}
	var channelInputParams *medialive.CreateChannelInput
	if err := json.Unmarshal(channel_template_raw, &channelInputParams); err != nil {
		logger.Error("Failed reading the template json to create a media channel. key: %v, err: %v", mediaKey, err)
		return "", err
	}
	rep := regexp.MustCompile("https://(.*)")
	fss := rep.FindStringSubmatch(*describeContainerOutput.Container.Endpoint)
	domain := fss[1]

	var destination_a = "mediastoressl://" + domain + "/LiveA/live"
	var destination_b = "mediastoressl://" + domain + "/LiveB/live"
	channelInputParams.Name = &mediaKey
	channelInputParams.InputAttachments[0].InputId = inputOutput.Input.Id
	channelInputParams.InputAttachments[0].InputAttachmentName = inputOutput.Input.Name
	channelInputParams.Destinations[0].Settings[0].Url = &destination_a
	channelInputParams.Destinations[0].Settings[1].Url = &destination_b
	// Create Channel
	channelOutput, err := ml.CreateChannel(channelInputParams)
	if err != nil {
		logger.Error("Failed creating a media channel. key: %v, err: %v", mediaKey, err)
		return "", err
	}
	// Get Store Container Result
	_, err = m.describeChannel(*channelOutput.Channel.Id)
	if err != nil {
		logger.Error("Failed creating a media channel. key: %v, err: %v", mediaKey, err)
		return "", err
	}
	url := *inputOutput.Input.Destinations[0].Url
	logger.Info("Suceeded creating a media channel. key: %v, endpoint: %v", mediaKey, url)

	// Store on the redis
	m.SetChannelOnRedis(mediaKey, *channelOutput.Channel.Id, url)

	return url, nil
}
