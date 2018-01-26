package godo

import (
	"sync"
)

type Task func() error

type pool struct {
	maxWorkers int
	taskRunner func(t Task)
	taskChan   chan Task
	errors     []error
	wg         *sync.WaitGroup
}

func NewPool(max int) *pool {
	p := &pool{
		maxWorkers: max,
		taskChan:   make(chan Task),
		wg:         &sync.WaitGroup{},
	}

	p.taskRunner = func(task Task) {
		err := task()

		if err != nil {
			p.errors = append(p.errors, err)
		}

		p.wg.Done()
	}

	p.startTaskListeners()

	return p
}

func (p *pool) Wait() bool {
	p.wg.Wait()

	return len(p.errors) == 0
}

func (p *pool) Errors() []error {
	return p.errors
}

func (p *pool) Do(task Task) {
	p.wg.Add(1)
	p.taskChan <- task
}

func (p *pool) startTaskListeners() {
	for i := 1; i <= p.maxWorkers; i++ {
		go func() {
			for {
				select {
				case task := <-p.taskChan:
					p.taskRunner(task)
				}
			}
		}()
	}
}
