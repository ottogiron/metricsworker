package processor

import (
	"context"
	"errors"
	"testing"
	"time"

	"reflect"

	"io/ioutil"
	"log"

	fworkerprocessor "github.com/ferrariframework/ferrariworker/processor"
	"github.com/ottogiron/metricsworker/worker"
)

var successfullJobs = []fworkerprocessor.Message{
	fworkerprocessor.Message{Payload: []byte("message 1")},
	fworkerprocessor.Message{Payload: []byte("message 2")},
	fworkerprocessor.Message{Payload: []byte("message 3")},
	fworkerprocessor.Message{Payload: []byte("message 4")},
	fworkerprocessor.Message{Payload: []byte("message 5")},
	fworkerprocessor.Message{Payload: []byte("message 6")},
}

type testMessagesHandler func(context context.Context) (<-chan fworkerprocessor.Message, error)

type processorAdapterMock struct {
	fworkerprocessor.Adapter
	handler  testMessagesHandler
	openErr  error
	closeErr error
}

func (s *processorAdapterMock) Open() error {
	if s.openErr != nil {
		return s.openErr
	}
	return nil
}

func (s *processorAdapterMock) Close() error {
	if s.closeErr != nil {
		return s.closeErr
	}
	return nil
}

func (s *processorAdapterMock) Messages(context context.Context) (<-chan fworkerprocessor.Message, error) {
	if s.handler == nil {
		return nil, errors.New("Please define a handler function for this mock handler")
	}
	return s.handler(context)
}

func mockMessagesHandler(messages []fworkerprocessor.Message) testMessagesHandler {
	return func(context context.Context) (<-chan fworkerprocessor.Message, error) {
		msgChannel := make(chan fworkerprocessor.Message)
		go func() {
			for _, message := range messages {
				msgChannel <- message
			}
			close(msgChannel)
		}()
		return msgChannel, nil
	}
}

func mockSleepMessagesHandler(duration time.Duration) testMessagesHandler {
	return func(context context.Context) (<-chan fworkerprocessor.Message, error) {
		msgChannel := make(chan fworkerprocessor.Message)
		go func() {
			time.Sleep(duration)
			msgChannel <- fworkerprocessor.Message{}
			close(msgChannel)

		}()
		return msgChannel, nil
	}
}

func newTestProcessor(adapter fworkerprocessor.Adapter) *processor {
	p := New(
		adapter,
		SetLogger(log.New(ioutil.Discard, "", 0)),
	)

	return p.(*processor)
}

func Test_processor_Start(t *testing.T) {
	type fields struct {
		adapter        fworkerprocessor.Adapter
		concurrency    int
		waitTimeout    time.Duration
		workerRegistry map[string]worker.Worker
	}

	logger := log.New(ioutil.Discard, "", 0)

	failedTaskError := errors.New("Failed task")
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"Process successful tasks",
			fields{
				&processorAdapterMock{
					handler: mockMessagesHandler(successfullJobs),
				},
				1,
				time.Millisecond * 200,
				map[string]worker.Worker{
					"distincName": &mockWorker{err: nil},
					"hourlyLog":   &mockWorker{err: failedTaskError},
				},
			},
			false,
		},
		{
			"Timed out",
			fields{
				&processorAdapterMock{
					handler: mockSleepMessagesHandler(time.Millisecond * 500),
				},
				1,
				200,
				map[string]worker.Worker{
					"distincName": &mockWorker{err: nil},
					"hourlyLog":   &mockWorker{err: failedTaskError},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(
				tt.fields.adapter,
				SetConcurrency(tt.fields.concurrency),
				SetWaitTimeout(tt.fields.waitTimeout),
				SetLogger(logger),
			)
			for id, worker := range tt.fields.workerRegistry {
				p.Register(id, worker)
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

			for _, wantResult := range tt.want {
				//find the wanted result in got results secuencially, since the same order cannot be guaranted
				for _, gotResult := range gotTasksResults {
					if wantResult.workerID == gotResult.workerID {
						if !reflect.DeepEqual(gotResult, wantResult) {
							t.Errorf("processor.Process() = %v want %v ", gotResult, wantResult)
						}
					}
				}
			}

		})
	}
}

type testWorkerHandler func(task interface{})

type mockWorker struct {
	err     error
	handler testWorkerHandler
}

func (mw *mockWorker) Execute(task interface{}) error {
	if mw.err != nil {
		return mw.err
	}

	if mw.handler != nil {
		mw.handler(task)
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
