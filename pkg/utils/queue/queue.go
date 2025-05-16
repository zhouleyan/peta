/*
 *  This file is part of PETA.
 *  Copyright (C) 2025 The PETA Authors.
 *  PETA is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Affero General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  PETA is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU Affero General Public License for more details.
 *
 *  You should have received a copy of the GNU Affero General Public License
 *  along with PETA. If not, see <https://www.gnu.org/licenses/>.
 */

package queue

import (
	"sync"
	"sync/atomic"
)

type Queue interface {
	Run()
	Push(job Job)
	Terminate()
}

/*
NewQueue
1. Make a queue
q := NewQueue(2, 10)

2. Run the queue
q.Run()

3. Push jobs
q.Push(job)

4. Terminate the queue
q.Terminate()
*/
func NewQueue(maxCapacity, maxThread int) *ChanQueue {
	return &ChanQueue{
		maxWorkers: maxThread,
		jobQueue:   make(chan Job, maxCapacity),
		workerPool: make(chan chan Job, maxThread),
		workers:    make([]*worker, maxThread),
		wg:         new(sync.WaitGroup),
	}
}

type ChanQueue struct {
	maxWorkers int
	jobQueue   chan Job
	workerPool chan chan Job
	workers    []*worker
	running    atomic.Bool
	wg         *sync.WaitGroup
}

func (q *ChanQueue) Run() {
	/*
		if *addr == old {
		    *addr = new
		    return true
		}
		return false
	*/

	// Ensure the queue is not started
	if !q.running.CompareAndSwap(false, true) {
		return
	}

	/*
		workerPool capacity: 2
	*/
	for i := 0; i < q.maxWorkers; i++ {
		q.workers[i] = newWorker(q.workerPool, q.wg)
		q.workers[i].Start()
	}

	go q.dispatch()
}

func (q *ChanQueue) dispatch() {
	// Pull the job from jobQueue
	for j := range q.jobQueue {
		// Pull the worker's idle jobChan
		jobChan := <-q.workerPool
		// Push the job to the worker's jobChan to exec
		jobChan <- j
	}
}

func (q *ChanQueue) Push(job Job) {
	// Ensure the queue is started
	if q.running.CompareAndSwap(false, true) {
		return
	}

	q.wg.Add(1)
	// Push job to Queue's jobQueue
	q.jobQueue <- job
}

func (q *ChanQueue) GetJobCount() int {
	return len(q.jobQueue)
}

func (q *ChanQueue) Terminate() {
	// Ensure the queue is started
	if q.running.CompareAndSwap(false, true) {
		return
	}

	// Wait all jobs are executed
	q.wg.Wait()

	close(q.jobQueue)
	for i := 0; i < q.maxWorkers; i++ {
		q.workers[i].Stop()
	}
	close(q.workerPool)
}
