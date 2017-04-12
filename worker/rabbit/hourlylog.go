package rabbit

import (
	"fmt"

	mgo "gopkg.in/mgo.v2"

	"time"

	"github.com/ottogiron/metricsworker/worker"
	"github.com/streadway/amqp"
)

var _ worker.Worker = (*DistinctNameWorker)(nil)

const eventsCollectionName = "hourly_events"

//HourlyLogWorker implementation of distinctname worker
type HourlyLogWorker struct {
	mongoHosts string
	dbName     string
}

//NewHourlyLogWorker returns a new instance of a distinctName worker
func NewHourlyLogWorker(eventsDB, mongoHosts string) *HourlyLogWorker {
	return &HourlyLogWorker{mongoHosts: mongoHosts, dbName: eventsDB}
}

//Execute executes a  DistinctNameWorker  task
func (w *HourlyLogWorker) Execute(task interface{}) error {
	delivery, ok := task.(amqp.Delivery)

	if !ok {
		return fmt.Errorf("Task should be a rabbit delivery %v", task)
	}

	countMetric, err := worker.UnmarshallCountMetric(delivery.Body)

	if err != nil {
		return fmt.Errorf("Failed to unmarshall rabbit delivery body (%s) %s ", string(delivery.Body), err)
	}
	now := time.Now().UTC()

	//I'm assuming the time in which the event happened is the rabbit delivery timestamp
	elapsed := now.Sub(delivery.Timestamp).Minutes()
	if elapsed <= 60 {
		session, err := mgo.Dial(w.mongoHosts)
		if err != nil {
			return fmt.Errorf("Dial to mongo servers failed %s %s", w.mongoHosts, err)
		}
		defer session.Close()
		c := session.DB(w.dbName).C(eventsCollectionName)
		err = c.Insert(countMetric)
		if err != nil {
			return fmt.Errorf("Failed to insert metric %s %v", err, countMetric)
		}
	}
	return nil
}
