package main

import (
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/tobiasbrodd/hivee-core/internal/client"
	"github.com/tobiasbrodd/hivee-core/internal/storage"
	"gopkg.in/yaml.v3"
)

type config struct {
	MQTT struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}
	Influx struct {
		Token string `yaml:"token"`
		Host  string `yaml:"host"`
		Port  int    `yaml:"port"`
	}
}

func (c *config) getConfig() *config {
	yamlFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Errorf("Config: %v", err.Error())
	}

	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Errorf("Config: %v", err.Error())
	}

	return c
}

func main() {
	keepAlive := make(chan os.Signal, 1)
	signal.Notify(keepAlive, os.Interrupt, syscall.SIGTERM)

	formatter := &log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05", FullTimestamp: true}
	log.SetFormatter(formatter)

	var c config
	c.getConfig()

	store := storage.New(c.Influx.Token, c.Influx.Host, c.Influx.Port, "Hivee")
	broker := client.New(c.MQTT.Host, c.MQTT.Port, "hivee-core", store)
	broker.Connect()

	<-keepAlive
	broker.Disconnect()
	store.Close()
}
