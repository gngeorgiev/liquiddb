package pool

import (
	"time"
)

type WorkerPool struct {
	size    int
	timeout time.Duration
	workers chan int
	work    chan func()
}

func NewWorkerPool(size int, timeout time.Duration) *WorkerPool {
	p := &WorkerPool{
		size:    size,
		timeout: timeout,
		workers: make(chan int, size),
		work:    make(chan func()),
	}

	for i := 0; i < size; i++ {
		p.workers <- i
	}

	go p.loop()

	return p
}

func (p *WorkerPool) loop() {
	for {
		select {
		case w := <-p.work:
			workerID := <-p.workers

			go func(workerID int) {
				done := make(chan struct{})

				go func() {
					w()
					done <- struct{}{}
				}()

				t := time.NewTimer(p.timeout)
				select {
				case <-t.C:
				case <-done:
				}

				t.Stop()

				p.workers <- workerID
			}(workerID)
		}
	}
}

func (p *WorkerPool) Schedule(work func()) {
	go func() {
		p.work <- work
	}()
}
