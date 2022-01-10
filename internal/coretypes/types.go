package coretypes

type AqaraTemp struct {
	Battery     float64 `json:"battery"`
	Humidity    float64 `json:"humidity"`
	Pressure    float64 `json:"pressure"`
	Temperature float64 `json:"temperature"`
	Voltage     int     `json:"voltage"`
	Linkquality int     `json:"linkquality"`
	Timestamp   int64   `json:"timestamp"`
}

type AqaraDoor struct {
	Battery     float64 `json:"battery"`
	Contact     bool    `json:"contact"`
	Temperature float64 `json:"temperature"`
	Voltage     int     `json:"voltage"`
	Linkquality int     `json:"linkquality"`
	Timestamp   int64   `json:"timestamp"`
}

type Measure struct {
	Value     interface{} `json:"value"`
	Timestamp int64       `json:"timestamp"`
	Location  string      `json:"location"`
}
