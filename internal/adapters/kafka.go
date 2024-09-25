package adapters

import (
	"context"
	"github.com/AydinKZ/K-Diode-Catcher/internal/domain"
)

type KafkaWriter struct {
	writer *kafka.Writer
}

func NewKafkaWriter(brokers []string, topic string) *KafkaWriter {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   topic,
	})
	return &KafkaWriter{writer: writer}
}

func (kw *KafkaWriter) WriteMessage(message domain.Message) error {
	return kw.writer.WriteMessages(context.Background(), kafka.Message{
		Topic: message.Topic,
		Value: []byte(message.Data),
	})
}
