package router

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
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

func buildAcceptKey(key string) string {
	h := sha1.New()
	h.Write([]byte(key))
	h.Write([]byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
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

	r.GET("/ws/observers", func(c *gin.Context) {
		fmt.Println(c.Request.Header.Get("Sec-Websocket-Key"))
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
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
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		hijacker := c.Writer.(http.Hijacker)
		_, readWriter, err := hijacker.Hijack()
		if err != nil {
			logger.Error("Failed to get connection: %v", err)
			return
		}
		key := c.Request.Header.Get("Sec-Websocket-Key")
		acceptKey := buildAcceptKey(key)

		readWriter.WriteString("HTTP/1.1 101 Switching Protocols\r\n")
		readWriter.WriteString("Upgrade: websocket\r\n")
		readWriter.WriteString("Connection: Upgrade\r\n")
		readWriter.WriteString("Sec-WebSocket-Accept: " + acceptKey + "\r\n")
		readWriter.WriteString("\r\n")
		readWriter.Flush()

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Error("Failed to get connection: %v", err)
			return
		}
		hub.RiderWorker(userId, h, conn, c)
	})
	return r
}
