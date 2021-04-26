package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	dt := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("%s Received message: %s from topic: %s\n", dt, msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection lost: %v", err)
}

func subscribe(client mqtt.Client, topic string) {
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s\n", topic)
}

func publish(client mqtt.Client, topic string, payload string) {
	token := client.Publish(topic, 1, false, payload)
	token.Wait()
}

func main() {
	keepAlive := make(chan os.Signal, 1)
	signal.Notify(keepAlive, os.Interrupt, syscall.SIGTERM)

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

	hivee := "hivee"
	subscribe(client, fmt.Sprintf("zigbee2mqtt/%s", hivee))

	<-keepAlive
	client.Disconnect(1000)
}
