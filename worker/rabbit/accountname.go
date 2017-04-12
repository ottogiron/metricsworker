package rabbit

import (
	"database/sql"
	"fmt"

	"time"

	"github.com/ottogiron/metricsworker/worker"
	"github.com/streadway/amqp"
)

var _ worker.Worker = (*AccountNameWorker)(nil)

//AccountNameWorker implementation of distinctname worker
type AccountNameWorker struct {
	db *sql.DB
}

//NewAccountNameWorker returns a new instance of a distinctName worker
func NewAccountNameWorker(db *sql.DB) *AccountNameWorker {

	return &AccountNameWorker{
		db: db,
	}
}

//Execute executes a  AccountNameWorker  task
func (w *AccountNameWorker) Execute(task interface{}) error {
	delivery, ok := task.(amqp.Delivery)

	if !ok {
		return fmt.Errorf("Task should be a rabbit delivery %v", task)
	}

	countMetric, err := worker.UnmarshallCountMetric(delivery.Body)

	if err != nil {
		return fmt.Errorf("Failed to unmarshall rabbit delivery body (%s) %s ", string(delivery.Body), err)
	}

	_, err = w.db.Exec(`
		INSERT INTO accounts ("username", "timestamp")
		Select CAST($1 AS VARCHAR), $2
		Where not exists (
		SELECT "username", "timestamp"
		FROM accounts
		WHERE username = $1
)
`, countMetric.UserName, time.Now().UTC().Unix())

	if err != nil {
		return fmt.Errorf("Failed to insert account username into database %s %s", countMetric.UserName, err)
	}
	return nil
}
