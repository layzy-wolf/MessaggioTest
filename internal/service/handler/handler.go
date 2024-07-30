package handler

import (
	"MessagioTest/config"
	"MessagioTest/pkg/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

type Consumer struct {
	db *gorm.DB
}

func (*Consumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (*Consumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (consumer *Consumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var message models.Message
		if err := json.Unmarshal(msg.Value, &message); err != nil {
			log.Printf("filed to unmarshall message: %v", err)
			continue
		}
		consumer.db.Create(&models.HandledMessage{
			From:    message.From,
			To:      message.To,
			Message: message.Message,
			Handled: true,
		})

		sess.MarkMessage(msg, "")
	}
	return nil
}

func initializeConsumerGroup(cfg *config.Cfg) (sarama.ConsumerGroup, error) {
	consumerGroup, err := sarama.NewConsumerGroup(
		[]string{cfg.Broker.Socket},
		cfg.Broker.CommonGroup,
		sarama.NewConfig(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize consumer group: %v", err)
	}

	return consumerGroup, nil
}

func SetupConsumerGroup(ctx context.Context, cfg *config.Cfg, db *gorm.DB) {
	consumerGroup, err := initializeConsumerGroup(cfg)
	if err != nil {
		log.Fatalf("initialization error: %v", err)
	}
	defer consumerGroup.Close()

	consumer := &Consumer{
		db: db,
	}

	for {
		err = consumerGroup.Consume(ctx, []string{cfg.Broker.CommonTopic}, consumer)
		if err != nil {
			log.Printf("error from consumer: %v", err)
		}
		if ctx.Err() != nil {
			return
		}
	}
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
	if err := db.AutoMigrate(&models.HandledMessage{}); err != nil {
		return nil, fmt.Errorf("error while migration: %v", err)
	}

	return db, nil
}
