package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var channels = make(map[string]func(mqtt.Client, string, []byte))
var sources = make(map[string]string)

type AqaraTemperature struct {
	Battery     float32 `json:"battery"`
	Humidity    float32 `json:"humidity"`
	Pressure    float32 `json:"pressure"`
	Temperature float32 `json:"temperature"`
	Voltage     int     `json:"voltage"`
	Timestamp   int64   `json:"timestamp"`
	Source      string  `json:"source"`
}

func handleAqaraTemperature(client mqtt.Client, topic string, payload []byte) {
	var message AqaraTemperature
	if err := json.Unmarshal(payload, &message); err != nil {
		log.Printf("Error: %v", err)
		return
	}

	message.Timestamp = time.Now().Unix()
	if src, ok := sources[topic]; ok {
		message.Source = src
	}

	payload, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	publish(client, "hivee/climate", payload)
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := msg.Payload()
	log.Printf("Received message: %s from topic: %s\n", payload, topic)

	if fn, ok := channels[topic]; ok {
		fn(client, topic, payload)
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connection lost: %v", err)
}

func subscribe(client mqtt.Client, topic string) {
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	log.Printf("Subscribed to topic: %s\n", topic)
}

func publish(client mqtt.Client, topic string, payload interface{}) {
	token := client.Publish(topic, 1, true, payload)
	token.Wait()
}

func main() {
	keepAlive := make(chan os.Signal, 1)
	signal.Notify(keepAlive, os.Interrupt, syscall.SIGTERM)

	channels["zigbee2mqtt/aqara_temperature"] = handleAqaraTemperature
	sources["zigbee2mqtt/aqara_temperature"] = "Indoor"

	broker := "localhost"
	port := 8883

	options := mqtt.NewClientOptions()
	options.AddBroker(fmt.Sprintf("mqtt://%s:%d", broker, port))
	options.SetClientID("hivee-core")
	options.SetDefaultPublishHandler(messagePubHandler)
	options.OnConnect = connectHandler
	options.OnConnectionLost = connectLostHandler

	client := mqtt.NewClient(options)
	token := client.Connect()

	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	for topic := range channels {
		subscribe(client, topic)
	}

	subscribe(client, "hivee/climate")

	<-keepAlive
	client.Disconnect(1000)
}
