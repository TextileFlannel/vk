package main

import (
	"fmt"
	"sync"
	"time"
)

type Worker struct {
	ID int
	Quit chan bool
}

type Pool struct {
	tasks chan string
	workers []*Worker
	wg sync.WaitGroup
	nextID int
	poolLock sync.Mutex
}


func NewPool() *Pool {
	return &Pool{
		tasks: make(chan string),
	}
}


func (w *Worker) Start(tasks <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case task, ok := <- tasks:
			if !ok {
				return
			}
			fmt.Printf("worker %d runing: %s\n", w.ID, task)
		case <-w.Quit:
			fmt.Printf("worker %d stoped\n", w.ID)
			return
		}
	}
}


func (w *Worker) Stop() {
	close(w.Quit)
}


func (p *Pool) AddWorker() {
	p.poolLock.Lock()
	defer p.poolLock.Unlock()

	worker := &Worker{
		ID: p.nextID,
		Quit: make(chan bool),
	}
	p.nextID += 1
	p.workers = append(p.workers, worker)
	p.wg.Add(1)
	go worker.Start(p.tasks, &p.wg)

	fmt.Printf("worker ID: %d\n", worker.ID)
}

func (p *Pool) RemoveWorker() {
	p.poolLock.Lock()
	defer p.poolLock.Unlock()

	if len(p.workers) == 0 {
		return
	}

	worker := p.workers[len(p.workers) - 1]
	p.workers = p.workers[:len(p.workers) - 1]
	worker.Stop()
	p.wg.Done()
	fmt.Printf("remove woker %d\n", worker.ID)
}


func main() {
	pool := NewPool()
	
	for i := 0; i < 3; i++ {
		pool.AddWorker()
	}

	for i := 0; i < 5; i++ {
		pool.tasks<-fmt.Sprintf("Task %d", i)
		time.Sleep(500 * time.Millisecond)
	}

	pool.AddWorker()
	time.Sleep(1 * time.Second)
	pool.RemoveWorker()

	for i := 5; i < 10; i++ {
		pool.tasks<-fmt.Sprintf("Task %d", i)
		time.Sleep(500 * time.Millisecond)
	}

	close(pool.tasks)
	pool.wg.Wait()
}