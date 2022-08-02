package media

import (
	"github.com/aws/aws-sdk-go/service/medialive"
)

func (m *Media) StopChannel(channelId string) error {
	ml := medialive.New(m.Sess)
	ml.StopChannel(&medialive.StopChannelInput{
		ChannelId: &channelId,
	})
	ml.WaitUntilChannelStopped(&medialive.DescribeChannelInput{
		ChannelId: &channelId,
	})
	return nil
}
