package application

import (
	"fmt"
	"github.com/AydinKZ/K-Diode-Catcher/config"
	"github.com/AydinKZ/K-Diode-Catcher/internal/adapters"
	"github.com/AydinKZ/K-Diode-Catcher/internal/domain"
	"github.com/AydinKZ/K-Diode-Catcher/internal/ports"
	"time"
)

type CatcherService struct {
	UDPReceiver    *adapters.UDPReceiver
	KafkaWriter    *adapters.KafkaWriter
	HashCalculator ports.MessageHashCalculator
	Cfg            config.Queue
	EnableHash     bool
}

func NewCatcherService(udpReceiver *adapters.UDPReceiver, kafkaWriter *adapters.KafkaWriter,
	calculator ports.MessageHashCalculator, cfg config.Queue, enableHash bool) *CatcherService {

	return &CatcherService{
		UDPReceiver:    udpReceiver,
		KafkaWriter:    kafkaWriter,
		HashCalculator: calculator,
		Cfg:            cfg,
		EnableHash:     enableHash,
	}
}

func (c *CatcherService) ReceiveAndPublishMessages() error {
	messageChan := make(chan domain.Message, 1000)
	defer close(messageChan)

	go func() {
		for msg := range messageChan {
			err := c.KafkaWriter.WriteMessage(msg)
			if err != nil {
				c.KafkaWriter.Log(fmt.Sprintf("[%v][Error Writing to Kafka] %v", time.Now(), err.Error()))
				continue
			}
			c.KafkaWriter.SendMetricsToKafka()
		}
	}()

	timeStart := time.Now()

	for {
		msg, err := c.UDPReceiver.Receive()
		if err != nil {
			adapters.BroadcastStatus(-2, msg.Topic, "ERROR", time.Since(timeStart))
			c.KafkaWriter.Log(fmt.Sprintf("[%v][Error Receiving UDP Message] %v", time.Now(), err.Error()))
			return err
		}

		if c.EnableHash {
			calculatedHash := c.HashCalculator.Calculate(msg.Value)
			if calculatedHash != msg.Hash {
				adapters.BroadcastStatusInc(-3, msg.Topic, "ERROR")
				c.KafkaWriter.Log(fmt.Sprintf("[%v][Error] %v, hash:%v, key:%v, value:%v", time.Now(), "hash mismatch", msg.Hash, msg.Key, msg.Value))
				return fmt.Errorf("hash mismatch")
			}
		}

		select {
		case messageChan <- msg:
			adapters.BroadcastStatus(0, msg.Topic, "SUCCESS", time.Since(timeStart))
		default:
			c.KafkaWriter.Log(fmt.Sprintf("[%v][Warning] Message channel is full, dropping message: %v", time.Now(), msg))
			adapters.BroadcastStatus(-4, msg.Topic, "DROPPED", time.Since(timeStart)) // if message channel is full, drop the message
		}
	}
}
