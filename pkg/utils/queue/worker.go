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

import "sync"

func newWorker(pool chan chan Job, wg *sync.WaitGroup) *worker {
	return &worker{
		pool:    pool,
		wg:      wg,
		jobChan: make(chan Job),
		quit:    make(chan struct{}),
	}
}

type worker struct {
	pool    chan chan Job
	wg      *sync.WaitGroup
	jobChan chan Job
	quit    chan struct{}
}

func (w *worker) Start() {
	// Make the worker's jobChan idle
	w.pool <- w.jobChan
	go w.Dispatch()
}

func (w *worker) Dispatch() {
	for {
		select {
		case j := <-w.jobChan:
			j.Exec()
			w.pool <- w.jobChan
			w.wg.Done()
		case <-w.quit:
			<-w.pool
			close(w.jobChan)
			return
		}
	}
}

func (w *worker) Stop() {
	close(w.quit)
}
