package worker

//Worker defines  a worker task processor
type Worker interface {
	Execute(task interface{}) error
}
