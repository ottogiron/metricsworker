package rabbit

import (
	"testing"

	"github.com/ottogiron/metricsworker/worker"
	"github.com/streadway/amqp"

	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func newMongoTestSession(t *testing.T) (*HourlyLogWorker, *mgo.Collection, func()) {
	host := "localhost"
	session, err := mgo.Dial(host)
	if err != nil {
		t.Fatalf("Dial to mongo servers failed %s %s", host, err)
	}
	db := "testDB"
	collection := session.DB(db).C(eventsCollectionName)
	w := NewHourlyLogWorker(db, host)
	return w, collection, func() {
		defer session.Close()
		session.DB(db).DropDatabase()
	}
}

func TestHourlyLogWorker_Execute(t *testing.T) {

	type args struct {
		task interface{}
	}
	tests := []struct {
		name string

		args    args
		wantErr bool
	}{
		{
			"Store hourly log in database",
			args{
				amqp.Delivery{
					Body:      validPayload,
					Timestamp: time.Now(),
				},
			},
			false,
		},
		{
			"Invalid payload",
			args{
				amqp.Delivery{
					Body:      invalidPayload,
					Timestamp: time.Now(),
				},
			},
			true,
		},
		{
			"Not a rabbit delivery",
			args{
				amqp.Delivery{
					Body:      invalidPayload,
					Timestamp: time.Now(),
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, collection, clean := newMongoTestSession(t)
			defer clean()
			err := w.Execute(tt.args.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("HourlyLogWorker.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			delivery, ok := tt.args.task.(amqp.Delivery)

			if !ok {
				t.Errorf("HourlyLogWorker.Execute() could not unmarshall delivery")
				return
			}

			metric, err := worker.UnmarshallCountMetric(delivery.Body)
			result := worker.CountMetric{}
			err = collection.Find(bson.M{"metric": metric.Metric}).One(&result)
			if err != nil {
				t.Errorf("HourlyLogWorker.Execute() could not find the stored metric")
			}
		})
	}
}
