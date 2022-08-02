package main

import (
	"flag"
	"fmt"
	"gps_logger/config"
	"gps_logger/logger"
	"gps_logger/media"
	"gps_logger/redis"
	"gps_logger/router"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

var (
	confPath = flag.String("c", "", "config file path (required).")
	conf     *config.Config
)

type LocationRequests []LocationRequest
type LocationRequest struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Timestamp int64   `json:"timestamp"`
}

func main() {
	flag.Parse()
	if *confPath == "" {
		flag.Usage()
		return
	}

	var err error
	conf, err = config.New(*confPath)
	if err != nil {
		fmt.Printf("[ERROR] Can not read configuration file: %v", err)
		return
	}

	logErr := logger.Init(conf)
	if logErr != nil {
		fmt.Println("[ERROR] Failed to initialize logger instance", err)
		return
	}

	cred := credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY"), os.Getenv("AWS_SECRET_ACCESS_KEY"), "")
	region := "ap-northeast-1"
	conf := &aws.Config{
		Credentials: cred,
		Region:      &region,
	}

	sess, awsErr := session.NewSession(conf)
	if awsErr != nil {
		logger.Error("Can not create new aws sessoin. error: %v", awsErr)
		return
	}

	rd, err := redis.New()
	if err != nil {
		logger.Error("Failed to create redis instance. error: %v", err)
		return
	}

	media, err := media.NewMediaLive(sess, rd)
	if err != nil {
		logger.Error("Failed to create media instance. error: %v", err)
		return
	}

	r := router.New(sess, rd, media)
	r.Run(":8080")
}
