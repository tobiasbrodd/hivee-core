package client

import (
	"encoding/json"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
	"github.com/tobiasbrodd/hivee-core/internal/storage"
)

type messageHandler func(mqtt.Client, string, []byte)

var channels = make(map[string]messageHandler)
var sources = make(map[string]string)
var store storage.Storage

type AqaraTemperatureSensor struct {
	Battery     float64 `json:"battery"`
	Humidity    float64 `json:"humidity"`
	Pressure    float64 `json:"pressure"`
	Temperature float64 `json:"temperature"`
	Voltage     int     `json:"voltage"`
	Timestamp   int64   `json:"timestamp"`
}

type Measurement struct {
	Value     float64 `json:"value"`
	Timestamp int64   `json:"timestamp"`
	Source    string  `json:"source"`
}

func publishMessage(mqttClient mqtt.Client, message interface{}, topic string) {
	payload, err := json.Marshal(message)
	if err != nil {
		log.Errorf("Cannot marshal: %v\n", err)
		return
	}

	publish(&mqttClient, fmt.Sprintf("hivee/%s", topic), payload, 0)
}

var handleAqaraTemperatureSensor messageHandler = func(mqttClient mqtt.Client, topic string, payload []byte) {
	var message AqaraTemperatureSensor
	if err := json.Unmarshal(payload, &message); err != nil {
		log.Errorf("Cannot unmarshal: %v\n", err)
		return
	}

	message.Timestamp = time.Now().Unix()
	src, ok := sources[topic]
	if !ok {
		src = "Unknown"
	}

	var temperature Measurement
	temperature.Value = message.Temperature
	temperature.Timestamp = message.Timestamp
	temperature.Source = src
	publishMessage(mqttClient, temperature, "temperature")
	store.StoreMeasure("temperature", temperature.Source, temperature.Value, temperature.Timestamp)

	var humidity Measurement
	humidity.Value = message.Humidity
	humidity.Timestamp = message.Timestamp
	humidity.Source = src
	publishMessage(mqttClient, humidity, "humidity")
	store.StoreMeasure("humidity", humidity.Source, humidity.Value, humidity.Timestamp)

	var pressure Measurement
	pressure.Value = message.Pressure
	pressure.Timestamp = message.Timestamp
	pressure.Source = src
	publishMessage(mqttClient, pressure, "pressure")
	store.StoreMeasure("pressure", pressure.Source, pressure.Value, pressure.Timestamp)
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

func Initialize(host string, port int, clientID string, initStore storage.Storage) *mqtt.Client {
	// mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	// mqtt.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
	// mqtt.WARN = log.New(os.Stdout, "[WARN] ", 0)
	// mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)

	store = initStore

	channels["zigbee2mqtt/aqara_temperature_sensor"] = handleAqaraTemperatureSensor
	sources["zigbee2mqtt/aqara_temperature_sensor"] = "Indoor"

	options := mqtt.NewClientOptions()
	options.AddBroker(fmt.Sprintf("mqtt://%s:%d", host, port))
	options.SetClientID(clientID)
	options.SetDefaultPublishHandler(messagePubHandler)
	options.OnConnect = connectHandler
	options.OnConnectionLost = connectLostHandler

	mqttClient := mqtt.NewClient(options)

	return &mqttClient
}

func Connect(mqttClient *mqtt.Client) mqtt.Token {
	token := (*mqttClient).Connect()

	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	for topic := range channels {
		subscribe(mqttClient, topic, 1)
	}

	subscribe(mqttClient, "hivee/temperature", 1)
	subscribe(mqttClient, "hivee/humidity", 1)
	subscribe(mqttClient, "hivee/pressure", 1)

	return token
}
