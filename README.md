# hivee-core

Core service for Hivee

## Guide

To start **hivee-core**, you can either run it directly using **Go**, run it in **Docker** or use **Docker Compose**.

### Go

```
go build
go run main.go
```

### Docker

```
docker build -t hivee-core .
docker run -it --rm hivee-core
```

### Docker Compose

```
docker-compose up -d --build
```

There are two services that need to be running before starting **hivee-core**. **mosquitto** as an MQTT broker and **InfluxDB** for storing time series data. Configurations for these should be in `config.yml`:
```
mqtt:
  host: <host.docker.internal>
  port: <1883>
influx:
  token: <generated by InfluxDB>
  host: <host.docker.internal>
  port: <8086>
```

### mosquitto

To start **mosquitto**, run:
```
mosquitto -c <path to config>
```

*macOS* path:
```
/usr/local/etc/mosquitto/<config name>.conf
```

*Linux* path:
```
/etc/mosquitto/conf.d/<config name>.conf
```

**mosquitto** can also be started as a service on Linux:
```
sudo systemctl start mosquitto
```

### InfluxDB

To start **InfluxDB**, run:
```
influxd
```

**InfluxDB** can also be started as a service on Linux:
```
sudo systemctl start influxdb
```
