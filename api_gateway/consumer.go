package main

import (
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
)

const (
	interval  = 7000
	noTimeout = -1
)

var topic = ""

type Consumer struct {
	consumer *kafka.Consumer
	stop     bool
}

func NewConsumer(topic string) (*Consumer, error) {
	con, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":        "localhost:9092",
		"group.id":                 "3",
		"auto.offset.reset":        "earliest",
		"session.timeout.ms":       interval,
		"enable.auto.offset.store": true,
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  5000,
	})
	if err != nil {
		return nil, err
	}

	if err = con.Subscribe(topic, nil); err != nil {
		return nil, err
	}

	err = con.Assign([]kafka.TopicPartition{
		{Topic: &topic, Partition: 0},
	})
	if err != nil {
		return nil, err
	}

	return &Consumer{consumer: con, stop: false}, nil
}

func (c *Consumer) Start(result chan Order) {
	for {
		if c.stop {
			break
		}

		kafkaMsg, err := c.consumer.ReadMessage(noTimeout)
		if err != nil {
			log.Printf("Consumer error: %v (%v)\n", err, kafkaMsg)
			continue
		}

		var order Order
		err = json.Unmarshal(kafkaMsg.Value, &order)
		if err != nil {
			log.Printf("Failed to unmarshal message: %s", err)
			continue
		}

		log.Printf("Get order: %v\n", order.OrderUID)

		result <- order
	}
}

func (c *Consumer) Stop() error {
	c.stop = true
	log.Println("Committed offset")
	return c.consumer.Close()
}
