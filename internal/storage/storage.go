package storage

import (
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	log "github.com/sirupsen/logrus"
)

type Storage struct {
	Client *influxdb2.Client
	writer *api.WriteAPI
}

func (storage Storage) StoreMeasure(measurement string, source string, value float64, timestamp int64) {
	log.Infof("Storing measurement %s: %f\n", measurement, value)
	p := influxdb2.NewPointWithMeasurement(measurement).
		AddTag("source", source).
		AddField("value", value).
		SetTime(time.Unix(timestamp, 0))
	(*storage.writer).WritePoint(p)
	(*storage.writer).Flush()
}

func Initialize(authToken string, host string, port int, org string, bucket string) (storage Storage) {
	client := influxdb2.NewClient(fmt.Sprintf("http://%s:%d", host, port), authToken)
	writer := client.WriteAPI(org, bucket)
	s := Storage{Client: &client, writer: &writer}
	errorsCh := writer.Errors()
	go func() {
		for err := range errorsCh {
			log.Errorf("InfluxDB: %s\n", err.Error())
		}
	}()

	return s
}
