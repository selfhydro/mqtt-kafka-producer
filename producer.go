package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/Shopify/sarama"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	kafkaConn = "localhost:9092"
	topic     = "sensor.temperature"
)

var producer sarama.SyncProducer

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
	publish(string(msg.Payload()), producer)
}

func main() {
	var err error
	producer, err = initProducer()
	if err != nil {
		fmt.Println("Error producer: ", err.Error())
		os.Exit(1)
	}

	opts := mqtt.NewClientOptions().AddBroker("localhost:1883").SetClientID("mqtt-kafka-producer")
	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if token := c.Subscribe("sensors/temperature", 0, f); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter msg: ")
		msg, _ := reader.ReadString('\n')

		publish(msg, producer)

		// publish with go routene
		// go publish(msg, producer)
	}

}

func initProducer() (sarama.SyncProducer, error) {
	sarama.Logger = log.New(os.Stdout, "", log.Ltime)

	config := sarama.NewConfig()
	config.Producer.Retry.Max = 5
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true

	// async producer
	// prd, err := sarama.NewAsyncProducer([]string{kafkaConn}, config)

	prd, err := sarama.NewSyncProducer([]string{kafkaConn}, config)

	return prd, err
}

func publish(message string, producer sarama.SyncProducer) {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}
	p, o, err := producer.SendMessage(msg)
	if err != nil {
		fmt.Println("Error publish: ", err.Error())
	}

	// publish async
	// producer.Input() <- msg

	fmt.Println("Partition: ", p)
	fmt.Println("Offset: ", o)
}
