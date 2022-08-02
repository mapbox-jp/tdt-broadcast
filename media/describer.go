package media

import (
	"time"

	"github.com/aws/aws-sdk-go/service/medialive"
	"github.com/aws/aws-sdk-go/service/mediastore"
)

type Channel struct {
	Id     string
	Url    string
	IsUsed bool
	UserId string
}

func (m *Media) describeContainer(containerName string) (*mediastore.DescribeContainerOutput, error) {
	var describeContainerOutput *mediastore.DescribeContainerOutput
	var err error
Loop:
	for {
		describeContainerOutput, err = m.Ms.DescribeContainer(&mediastore.DescribeContainerInput{
			ContainerName: &containerName,
		})
		if err != nil {
			return nil, err
		} else if *describeContainerOutput.Container.Status == "ACTIVE" {
			break Loop
		}
		time.Sleep(time.Second * 1)
	}
	return describeContainerOutput, nil
}

func (m *Media) describeChannel(channelId string) (*medialive.DescribeChannelOutput, error) {
	var describeChannelOutput *medialive.DescribeChannelOutput
	var err error
Loop:
	for {
		describeChannelOutput, err = m.Ml.DescribeChannel(&medialive.DescribeChannelInput{
			ChannelId: &channelId,
		})
		if err != nil {
			return nil, err
		} else if *describeChannelOutput.State == "IDLE" {
			break Loop
		}
		time.Sleep(time.Second * 1)
	}
	return describeChannelOutput, nil
}
