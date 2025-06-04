package util

import (
	"runtime"
)

type (
	job[T any]        func(data T)
	workerPool[T any] struct {
		poolSize int
		ch       chan T
		job      job[T]
	}
)

const defaultPoolSize = 5

func NewWorkerPool[T any](f job[T], poolSize ...int) *workerPool[T] {
	w := &workerPool[T]{
		job: f,
	}
	if len(poolSize) == 0 {
		poolSize = []int{defaultPoolSize}
	}
	w.poolSize = poolSize[0]

	// https://github.com/valyala/fasthttp/blob/master/workerpool.go
	if runtime.GOMAXPROCS(0) == 1 {
		w.ch = make(chan T)
	} else {
		w.ch = make(chan T, 1)
	}

	for i := 0; i < w.poolSize; i++ {
		go func() {
			for data := range w.ch {
				w.job(data)
			}
		}()
	}

	return w
}

func (w *workerPool[T]) Do(data T) {
	w.ch <- data
}

func (w *workerPool[T]) Close() {
	close(w.ch)
}
