package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection lost: %v", err)
}

func getTLSConfig() *tls.Config {
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("mosquitto.org.crt")
	if err != nil {
		panic(err.Error())
	}

	certPool.AppendCertsFromPEM(ca)
	return &tls.Config{RootCAs: certPool}
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
	broker := "127.0.0.1"
	port := 1883

	options := mqtt.NewClientOptions()
	options.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	options.SetClientID("hivee_core")
	options.SetDefaultPublishHandler(messagePubHandler)
	options.OnConnect = connectHandler
	options.OnConnectionLost = connectLostHandler
	options.SetTLSConfig(getTLSConfig())

	client := mqtt.NewClient(options)
	token := client.Connect()

	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	topic := "topic/hivee"
	subscribe(client, topic)

	for i := 0; i < 5; i++ {
		payload := fmt.Sprintf("Payload %d", i)
		publish(client, topic, payload)
		time.Sleep(time.Second)
	}

	client.Disconnect(1000)
}
