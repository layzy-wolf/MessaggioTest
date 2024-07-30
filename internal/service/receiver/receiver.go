package receiver

import (
	"MessagioTest/config"
	"MessagioTest/pkg/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"sync"
)

var (
	mu            sync.Mutex
	messagesQueue chan *models.Message
)

func SetupProducer(cfg *config.Cfg) (sarama.SyncProducer, error) {
	brokerConfig := sarama.NewConfig()
	brokerConfig.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer([]string{cfg.Broker.Socket}, brokerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to setup producer: %w", err)
	}
	return producer, nil
}

func SetupSender(ctx context.Context, messages <-chan *models.Message, producer sarama.SyncProducer, cfg *config.Cfg) {
	for {
		select {
		case msg, ok := <-messages:
			if ok {
				if err := sendMessage(msg, producer, cfg); err != nil {
					return
				}
				continue
			}
			return
		case <-ctx.Done():
			return
		}
	}
}

func sendMessage(msg *models.Message, producer sarama.SyncProducer, cfg *config.Cfg) error {
	mu.Lock()
	defer mu.Unlock()
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	m := &sarama.ProducerMessage{
		Topic: cfg.Broker.CommonTopic,
		Value: sarama.StringEncoder(msgJSON),
	}

	_, _, err = producer.SendMessage(m)
	if err != nil {
		return fmt.Errorf("error while send message to broker: %w", err)
	}
	return nil
}

func SetupReceiver() (chan *models.Message, func()) {
	messagesQueue = make(chan *models.Message)
	return messagesQueue, func() { close(messagesQueue) }
}

func SetupDB(cfg *config.Cfg) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.DB.Host,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.DB,
		cfg.DB.Port,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error while creating connection to db: %v", err)
	}
	if err := db.AutoMigrate(&models.Message{}); err != nil {
		return nil, fmt.Errorf("error while migration: %v", err)
	}

	return db, nil
}

func SetupAPIReceiver(db *gorm.DB, messages chan<- *models.Message) gin.HandlerFunc {
	return func(c *gin.Context) {
		var msg models.Message
		if err := c.BindJSON(&msg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "incorrect json message: " + err.Error()})
			return
		}
		mu.Lock()
		messages <- &msg
		mu.Unlock()

		res := db.Create(&msg)

		if res.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "something wrong while insert to db: " + res.Error.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("message pass to handler, id is %v", msg.ID)})
	}
}

func GetStatistic(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var messageCount, handledMessageCount int64
		res, handledRes := db.Model(models.Message{}).Count(&messageCount),
			db.Model(models.HandledMessage{}).Count(&handledMessageCount)

		if res.Error != nil && handledRes.Error != nil {
			c.JSON(http.StatusInternalServerError,
				gin.H{
					"message": fmt.Sprintf("error while querying db: %v, %v", res.Error.Error(), handledRes.Error.Error()),
				},
			)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("percentage of handled messages: %v%%", (float64(handledMessageCount)/float64(messageCount))*100)})
	}
}
