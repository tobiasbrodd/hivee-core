package client

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
	"github.com/tobiasbrodd/hivee-core/internal/coretypes"
	"github.com/tobiasbrodd/hivee-core/internal/storage"
)

type Client struct {
	MQTT *mqtt.Client
}

type messageHandler func(mqtt.Client, string, []byte)

var channels = make(map[string]messageHandler)
var locations = make(map[string]string)
var store *storage.Storage

func publishMessage(mqttClient mqtt.Client, message interface{}, topic string) {
	payload, err := json.Marshal(message)
	if err != nil {
		log.Errorf("Cannot marshal: %v\n", err)
		return
	}

	publish(&mqttClient, fmt.Sprintf("hivee/%s", topic), payload, 0)
}

func getSource(topic string) string {
	srcs := strings.Split(topic, "/")
	return srcs[len(srcs)-1]
}

var handleAqaraMeasure messageHandler = func(mqttClient mqtt.Client, topic string, payload []byte) {
	var message coretypes.AqaraMeasure
	if err := json.Unmarshal(payload, &message); err != nil {
		log.Errorf("Cannot unmarshal: %v\n", err)
		return
	}

	message.Timestamp = time.Now().Unix()

	src := getSource(topic)
	store.StoreAqaraMeasure(src, message)

	location, ok := locations[topic]
	if !ok {
		location = "Unknown"
	}

	var temperature coretypes.Measure
	temperature.Value = message.Temperature
	temperature.Timestamp = message.Timestamp
	temperature.Location = location
	publishMessage(mqttClient, temperature, "temperature")
	store.StoreMeasure("temperature", temperature)

	var humidity coretypes.Measure
	humidity.Value = message.Humidity
	humidity.Timestamp = message.Timestamp
	humidity.Location = location
	publishMessage(mqttClient, humidity, "humidity")
	store.StoreMeasure("humidity", humidity)

	var pressure coretypes.Measure
	pressure.Value = message.Pressure
	pressure.Timestamp = message.Timestamp
	pressure.Location = location
	publishMessage(mqttClient, pressure, "pressure")
	store.StoreMeasure("pressure", pressure)
}

var messagePubHandler mqtt.MessageHandler = func(mqttClient mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := msg.Payload()
	log.Infof("Received message: %s from topic: %s\n", payload, topic)

	if fn, ok := channels[topic]; ok {
		fn(mqttClient, topic, payload)
	}
}

var connectHandler mqtt.OnConnectHandler = func(mqttClient mqtt.Client) {
	log.Info("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(mqttClient mqtt.Client, err error) {
	log.Errorf("Connection lost: %v\n", err)
}

func subscribe(mqttClient *mqtt.Client, topic string, qos byte) {
	token := (*mqttClient).Subscribe(topic, qos, nil)
	go func() {
		_ = token.Wait()
		if token.Error() != nil {
			log.Errorf("MQTT: %v\n", token.Error())
		}
	}()
	log.Infof("Subscribed to topic: %s\n", topic)
}

func publish(mqttClient *mqtt.Client, topic string, payload []byte, qos byte) {
	token := (*mqttClient).Publish(topic, qos, true, payload)
	go func() {
		_ = token.Wait()
		if token.Error() != nil {
			log.Printf("MQTT error: %v\n", token.Error())
		}
	}()
	log.Infof("Published to topic: %s\n", topic)
}

func New(host string, port int, clientID string, initStore *storage.Storage) Client {
	// mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	// mqtt.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
	// mqtt.WARN = log.New(os.Stdout, "[WARN] ", 0)
	// mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)

	store = initStore

	channels["zigbee2mqtt/aqara_temperature"] = handleAqaraMeasure
	locations["zigbee2mqtt/aqara_temperature"] = "Indoor"

	options := mqtt.NewClientOptions()
	options.AddBroker(fmt.Sprintf("mqtt://%s:%d", host, port))
	options.SetClientID(clientID)
	options.SetDefaultPublishHandler(messagePubHandler)
	options.OnConnect = connectHandler
	options.OnConnectionLost = connectLostHandler

	mqttClient := mqtt.NewClient(options)
	client := Client{MQTT: &mqttClient}

	return client
}

func (client Client) Connect() mqtt.Token {
	token := (*client.MQTT).Connect()

	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	for topic := range channels {
		subscribe(client.MQTT, topic, 1)
	}

	subscribe(client.MQTT, "hivee/temperature", 1)
	subscribe(client.MQTT, "hivee/humidity", 1)
	subscribe(client.MQTT, "hivee/pressure", 1)

	return token
}

func (client Client) Disconnect() {
	(*client.MQTT).Disconnect(1000)
}
