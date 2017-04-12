package rabbit

import (
	"testing"

	"time"

	"github.com/go-redis/redis"
	"github.com/ottogiron/metricsworker/worker"
	"github.com/streadway/amqp"
)

func testRedisClient(t *testing.T) (*redis.Client, func()) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       2,  // use default DB
	})

	_, err := client.Ping().Result()
	if err != nil {
		t.Fatalf("Failed to connect to redis %s", err)
	}
	return client, func() {
		client.FlushDb()
	}
}

func TestDistinctNameWorker_Execute(t *testing.T) {
	client, clean := testRedisClient(t)

	type fields struct {
		rclient *redis.Client
	}
	type args struct {
		task interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Store EVent Succesfuly",
			fields{client},
			args{
				amqp.Delivery{
					Body: validPayload,
				},
			},
			false,
		},
		{
			"Fail to unmarshall",
			fields{client},
			args{
				amqp.Delivery{
					Body: invalidPayload,
				},
			},
			true,
		},
		{
			"Not a rabbit amqp delivery",
			fields{client},
			args{
				nil,
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer clean()
			w := NewDistincNameWorker(tt.fields.rclient)
			err := w.Execute(tt.args.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("DistinctNameWorker.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return

			}

			if err != nil {
				return
			}

			unixTimeNow := time.Now().UTC().Unix()
			zSlice := client.ZRangeWithScores("events", 0, unixTimeNow+500)

			delivery, ok := tt.args.task.(amqp.Delivery)

			if !ok {
				t.Errorf("DistinctNameWorker.Execute() could not unmarshall delivery")
				return
			}

			metric, err := worker.UnmarshallCountMetric(delivery.Body)
			if err != nil {
				t.Errorf("DistinctNameWorker.Execute() could not unmarshall count metric")
			}
			for _, z := range zSlice.Val() {
				val := client.HGet(z.Member.(string), "metric").Val()
				if val != metric.Metric {
					t.Errorf("DistincNameWorker.Execute() stored metric = %s want %v", val, metric.Metric)
				}
			}
		})
	}
}
