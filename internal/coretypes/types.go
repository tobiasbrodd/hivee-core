package coretypes

type AqaraMeasure struct {
	Battery     float64 `json:"battery"`
	Humidity    float64 `json:"humidity"`
	Pressure    float64 `json:"pressure"`
	Temperature float64 `json:"temperature"`
	Voltage     int     `json:"voltage"`
	Timestamp   int64   `json:"timestamp"`
}

type Measure struct {
	Value     float64 `json:"value"`
	Timestamp int64   `json:"timestamp"`
	Location  string  `json:"location"`
}
