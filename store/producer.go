package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const (
	flushTimeout = 7000
)

type Producer struct {
	producer *kafka.Producer
}

func NewProd() (*Producer, error) {
	pro, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092"})
	if err != nil {
		return nil, fmt.Errorf("error creating producer: %v", err)
	}

	return &Producer{producer: pro}, nil

}

func (p *Producer) Close() {
	p.producer.Flush(flushTimeout)
	p.producer.Close()
}

func (p *Producer) Produce(message []byte, topic string) error {
	kfkMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          message,
		Key:            nil,
	}

	kfkChan := make(chan kafka.Event)
	if err := p.producer.Produce(kfkMsg, kfkChan); err != nil {
		return fmt.Errorf("error sending message to kafka: %v", err)
	}

	e := <-kfkChan
	switch ev := e.(type) {
	case *kafka.Message:
		return nil
	case kafka.Error:
		return ev
	default:
		return fmt.Errorf("unknown message type: %v", ev)
	}
}
