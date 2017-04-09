package processor

import (
	"context"
	"errors"
	"testing"
	"time"

	"reflect"

	fworkerprocessor "github.com/ferrariframework/ferrariworker/processor"
	"github.com/ottogiron/metricsworker/worker"
)

type processorAdapterMock struct {
	tb       testing.TB
	messages []fworkerprocessor.Message
}

func (s *processorAdapterMock) Open() error {
	return nil
}

func (s *processorAdapterMock) Close() error {
	return nil
}

func (s *processorAdapterMock) Messages(context context.Context) (<-chan fworkerprocessor.Message, error) {
	msgChannel := make(chan fworkerprocessor.Message)
	go func() {
		for _, message := range s.messages {
			msgChannel <- message
		}
	}()
	return msgChannel, nil
}

func newTestProcessor(adapter fworkerprocessor.Adapter) *processor {
	p := NewProcessor(adapter)
	return p.(*processor)
}

func Test_processor_Start(t *testing.T) {
	type fields struct {
		adapter        fworkerprocessor.Adapter
		concurrency    int
		waitTimeout    time.Duration
		workerRegistry map[string]worker.Worker
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &processor{
				adapter:        tt.fields.adapter,
				concurrency:    tt.fields.concurrency,
				waitTimeout:    tt.fields.waitTimeout,
				workerRegistry: tt.fields.workerRegistry,
			}
			if err := p.Start(); (err != nil) != tt.wantErr {
				t.Errorf("processor.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_processor_handleFailedTask(t *testing.T) {
	type fields struct {
		adapter        fworkerprocessor.Adapter
		concurrency    int
		waitTimeout    time.Duration
		workerRegistry map[string]worker.Worker
	}
	type args struct {
		taskResult *taskResult
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &processor{
				adapter:        tt.fields.adapter,
				concurrency:    tt.fields.concurrency,
				waitTimeout:    tt.fields.waitTimeout,
				workerRegistry: tt.fields.workerRegistry,
			}
			p.handleFailedTask(tt.args.taskResult)
		})
	}
}

func Test_processor_process(t *testing.T) {
	type fields struct {
		workerRegistry map[string]worker.Worker
	}
	type args struct {
		task       interface{}
		workersIDS []string
	}

	failedTaskError := errors.New("Failed task")

	tests := []struct {
		name   string
		fields fields
		args   args
		want   []taskResult
	}{
		{
			"Process sucessfuly",
			fields{
				workerRegistry: map[string]worker.Worker{
					"distincName": &mockWorker{err: nil},
					"hourlyLog":   &mockWorker{err: failedTaskError},
				},
			},
			args{
				task:       "simple task value",
				workersIDS: []string{"distincName", "hourlyLog"},
			},
			[]taskResult{
				taskResult{
					workerID: "distincName",
				},
				taskResult{
					workerID: "hourlyLog",
					err:      failedTaskError,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestProcessor(nil)
			p.workerRegistry = tt.fields.workerRegistry
			got := p.process(tt.args.task, tt.args.workersIDS...)

			gotTasksResults := []taskResult{}
			for gotTaskResult := range got {
				gotTasksResults = append(gotTasksResults, gotTaskResult)
			}

			if !reflect.DeepEqual(gotTasksResults, tt.want) {
				t.Errorf("processor.Process() = %v want %v ", gotTasksResults, tt.want)
			}

		})
	}
}

func testTasksResultChan(taskResults ...taskResult) <-chan taskResult {
	out := make(chan taskResult)
	go func() {
		for _, taskResult := range taskResults {
			out <- taskResult
		}
		close(out)
	}()

	return out
}

type mockWorker struct {
	err error
}

func (mw *mockWorker) Execute(task interface{}) error {
	if mw.err != nil {
		return mw.err
	}
	return nil
}
func Test_processor_Register(t *testing.T) {

	type args struct {
		id     string
		worker worker.Worker
	}
	tests := []struct {
		name string
		args args
	}{
		{"Register distincName worker", args{"disctincName", &mockWorker{}}},
		{"Register hourly log worker", args{"hourlyLog", &mockWorker{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestProcessor(nil)
			p.Register(tt.args.id, tt.args.worker)

			if p.workerRegistry[tt.args.id] != tt.args.worker {
				t.Errorf("processor.Register() expected worker to be registered")
			}
		})
	}
}
