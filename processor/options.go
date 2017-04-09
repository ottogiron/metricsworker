package processor

import "time"

//Option a functional option for the processor
type Option func(*processor)

//SetConcurrency sets the concurrency for registered workers
func SetConcurrency(concurrency int) Option {
	return func(p *processor) {
		p.concurrency = concurrency
	}
}

//SetWaitTimeout sets the time the processor will wait (keep the connection open) for new tasks
func SetWaitTimeout(waitTimeout time.Duration) Option {
	return func(p *processor) {
		p.waitTimeout = waitTimeout
	}
}
