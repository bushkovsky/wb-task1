package main

import (
	"database/sql"
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
)

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

func (c *Consumer) Start(db *sql.DB) {
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

		err = saveOrder(db, order)
		if err != nil {
			log.Printf("Failed to save order to database: %s", err)
			continue
		}

		log.Printf("Order saved: %v\n", order.OrderUID)

		if _, err = c.consumer.StoreMessage(kafkaMsg); err != nil {
			log.Println(err)
			continue
		}
	}
}

func (c *Consumer) StartToSelect(prod *Producer, db *sql.DB) {
	for {
		if c.stop {
			break
		}

		kafkaMsg, err := c.consumer.ReadMessage(noTimeout)
		if err != nil {
			log.Printf("Consumer error: %v (%v)\n", err, kafkaMsg)
			continue
		}

		var id OrderID
		err = json.Unmarshal(kafkaMsg.Value, &id)
		if err != nil {
			log.Printf("Failed to unmarshal message: %s", err)
			continue
		}
		var ord *Order
		ord, err = getOrderByUID(db, id.OrderUID)
		if err != nil {
			log.Printf("Failed to selected order to database: %s", err)
			continue
		}

		message, err := json.Marshal(ord)
		if err != nil {
			log.Printf("Failed to encode order to JSON")
			return
		}

		log.Printf("Order has selected: %v\n", ord.OrderUID)
		err = prod.Produce(message, topic3)
		if err != nil {
			log.Printf("Can not send response to kafka")
		}

		if _, err = c.consumer.StoreMessage(kafkaMsg); err != nil {
			log.Println(err)
			continue
		}
	}
}
