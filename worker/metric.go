package worker

//CountMetric represents a count metric of different types of events
type CountMetric struct {
	UserName string `json:"username"`
	Count    int64  `json:"count"`
	Metric   string `json:"kite_call"`
}
