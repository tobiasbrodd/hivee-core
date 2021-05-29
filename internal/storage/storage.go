package storage

import (
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	log "github.com/sirupsen/logrus"
	"github.com/tobiasbrodd/hivee-core/internal/coretypes"
)

type Storage struct {
	Influx       *influxdb2.Client
	Organization string
}

func (storage *Storage) StoreAqaraMeasure(measurement string, measure coretypes.AqaraMeasure) {
	log.Infof("Storing measurement %s\n", measurement)

	writer := storage.getWriter("pibee")
	p := influxdb2.NewPointWithMeasurement(measurement).
		AddField("battery", measure.Battery).
		AddField("humidity", measure.Humidity).
		AddField("pressure", measure.Pressure).
		AddField("temperature", measure.Temperature).
		AddField("voltage", measure.Voltage).
		SetTime(time.Unix(measure.Timestamp, 0))
	(*writer).WritePoint(p)
	(*writer).Flush()
}

func (storage *Storage) StoreMeasure(measurement string, measure coretypes.Measure) {
	log.Infof("Storing measurement %s\n", measurement)

	writer := storage.getWriter("hivee")
	p := influxdb2.NewPointWithMeasurement(measurement).
		AddTag("location", measure.Location).
		AddField("value", measure.Value).
		SetTime(time.Unix(measure.Timestamp, 0))
	(*writer).WritePoint(p)
	(*writer).Flush()
}

func (storage *Storage) getWriter(bucket string) *api.WriteAPI {
	writer := (*storage.Influx).WriteAPI(storage.Organization, bucket)
	errorsCh := writer.Errors()
	go func() {
		for err := range errorsCh {
			log.Errorf("InfluxDB: %s\n", err.Error())
		}
	}()

	return &writer
}

func New(authToken string, host string, port int, org string) *Storage {
	client := influxdb2.NewClient(fmt.Sprintf("http://%s:%d", host, port), authToken)
	storage := &Storage{Influx: &client, Organization: org}

	return storage
}

func (storage Storage) Close() {
	(*storage.Influx).Close()
}
