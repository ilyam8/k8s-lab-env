package bouncer

import (
	"sync"
	"time"
)

type task struct {
	once    sync.Once
	done    chan struct{}
	running chan struct{}
}

func (t *task) stop() chan struct{} {
	t.once.Do(func() { close(t.done) })
	return t.running
}

func newTask(doWork func(), doEvery time.Duration) *task {
	task := task{
		done:    make(chan struct{}),
		running: make(chan struct{}),
	}

	go func() {
		t := time.NewTicker(doEvery)
		defer func() {
			t.Stop()
			close(task.running)
		}()
		for {
			select {
			case <-task.done:
				return
			case <-t.C:
				doWork()
			}
		}
	}()

	return &task
}
