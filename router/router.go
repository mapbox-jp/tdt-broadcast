package router

import (
	"gps_logger/hub"
	"gps_logger/logger"
	"gps_logger/media"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func New(sess *session.Session, rd *redis.Client, media media.MediaRepository) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	if os.Getenv("ENV") == "dev" {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost",
		},
		AllowMethods: []string{
			"GET",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Access-Control-Allow-Credentials",
			"Access-Control-Allow-Headers",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"Authorization",
			"X-Requested-With",
		},
		AllowCredentials: true,
		MaxAge:           24 * time.Hour,
	}))

	h := hub.NewHub(rd, media)
	go h.Run()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"alive": true,
		})
	})

	r.GET("/rider", func(c *gin.Context) {
		r.LoadHTMLFiles("rider.html")
		c.HTML(200, "rider.html", nil)
	})

	r.GET("/observer", func(c *gin.Context) {
		r.LoadHTMLFiles("observer.html")
		c.HTML(200, "observer.html", nil)
	})

	r.POST("/channels", func(c *gin.Context) {
		endpoint, err := media.CreateChannel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"err": err,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"url": endpoint,
			})
		}
	})

	r.PUT("/channels/:mediaKey/:channelId", func(c *gin.Context) {
		var err error
		mediaKey := c.Param("mediaKey")
		channelId := c.Param("channelId")
		type ChannelBody struct {
			IsUsed bool   `json:"is_used"`
			UserId string `json:"user_id"`
		}
		var body ChannelBody
		if err = c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if !body.IsUsed {
			err = media.StopChannel(mediaKey, channelId)
		} else {
			err = media.StartChannel(body.UserId, mediaKey, channelId)
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"err": err,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{})
		}
	})

	r.GET("/ws/observe/riders", func(c *gin.Context) {
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Error("Failed to get connection: %v", err)
			return
		}
		hub.ObserverWorker(h, conn, c)
	})

	r.GET("/ws/riders", func(c *gin.Context) {
		userId := c.Param("user_id")
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Error("Failed to get connection: %v", err)
			return
		}
		hub.RiderWorker(userId, h, conn, c)
	})
	return r
}
