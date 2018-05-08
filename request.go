package main

import (
	"container/heap"
)

type Request struct {
	From string
	Clock
	failed  bool
	inquire *InterNodeMessage
}

type RequestQ []Request

func NewRequest(from string) Request {
	localClock.Inc()
	return Request{
		From:  from,
		Clock: localClock,
	}
}

func NewRequestQ() *RequestQ {
	r := make(RequestQ, 0)
	heap.Init(&r)

	return &r
}

func (self RequestQ) Len() int { return len(self) }
func (self RequestQ) Less(i, j int) bool {
	if self[i].Clock == self[j].Clock {
		return self[i].From < self[j].From
	}
	return self[i].Clock < self[j].Clock
}
func (self RequestQ) Swap(i, j int)       { self[i], self[j] = self[j], self[i] }
func (self *RequestQ) Push(x interface{}) { *self = append(*self, x.(Request)) }
func (self *RequestQ) Pop() (popped interface{}) {
	popped = self.Peek()
	*self = (*self)[:len(*self)-1]
	return
}

func (self *RequestQ) Peek() (popped interface{}) {
	popped = (*self)[len(*self)-1]
	return
}

func Less(r Request, msg InterNodeMessage) bool {
	if r.Clock == msg.Clock {
		return r.From < msg.From
	}
	return r.Clock < msg.Clock
}
