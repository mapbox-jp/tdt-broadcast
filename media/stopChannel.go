package media

import (
	"github.com/aws/aws-sdk-go/service/medialive"
)

func (m *Media) StopChannel(mediaKey string, channelId string) error {
	err := m.UpdateStoppedChannelOnRedis(mediaKey)
	if err != nil {
		return err
	}
	ml := medialive.New(m.Sess)
	_, err = ml.StopChannel(&medialive.StopChannelInput{
		ChannelId: &channelId,
	})
	if err != nil {
		return err
	}
	err = ml.WaitUntilChannelStopped(&medialive.DescribeChannelInput{
		ChannelId: &channelId,
	})
	return err
}
