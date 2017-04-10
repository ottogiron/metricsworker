package rabbit

import "github.com/ottogiron/metricsworker/worker"
import "fmt"

var _ worker.Worker = (*DistinctNameWorker)(nil)

//DistinctNameWorker implementation of distinctname worker
type DistinctNameWorker struct {
}

//Execute executes a  DistinctNameWorker  task
func (w *DistinctNameWorker) Execute(task interface{}) error {
	fmt.Print(task)
	return nil
}
