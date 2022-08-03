package media

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/medialive"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/go-redis/redis"
)

type Media struct {
	Ml   *medialive.MediaLive
	Ms   *mediastore.MediaStore
	Rd   *redis.Client
	Sess *session.Session
}

type MediaRepository interface {
	CreateChannel() (string, error)
	GetChannel() (string, string, string, error)
	GetStore(mediaKey string) (string, error)
	describeContainer(containerName string) (*mediastore.DescribeContainerOutput, error)
	describeChannel(channelId string) (*medialive.DescribeChannelOutput, error)
	StartChannel(userId string, mediaKey string, channelId string) error
	StopChannel(mediaKey string, channelId string) error
}

func NewMediaLive(sess *session.Session, rd *redis.Client) (MediaRepository, error) {
	ml := medialive.New(sess)
	ms := mediastore.New(sess)
	return &Media{
		Ml:   ml,
		Ms:   ms,
		Rd:   rd,
		Sess: sess,
	}, nil
}
