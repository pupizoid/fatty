package lib

import (
	"fmt"
	"os"
	"os/signal"
	"github.com/fatih/color"
	"sync"
	"time"
	"net/url"
)

type Dispatcher struct {
	stats *RunStats
	Proxy *url.URL

	Emitters []*Emitter

	log chan LogMessage

	stop, done chan struct{}
}

func NewDispatcher() *Dispatcher {
	stats := &RunStats{counter: &RequestCounter{mutex: &sync.RWMutex{}}}

	done := make(chan struct{})
	stop := make(chan struct{})

	return &Dispatcher{stats: stats, done: done, stop: stop}
}

func (d *Dispatcher) Run() {

	nthreads := len(d.Emitters)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	d.log = make(chan LogMessage, nthreads * 10)

	for _, e := range d.Emitters {
		go e.Start(d.stats.counter, d.Proxy, d.stop, d.done, d.log)
	}

	for nthreads > 0 {
		select {
		case msg := <-d.log:
			if msg.err {
				fmt.Println(color.RedString("Returned request with err code: %d, message: %s", msg.ReqCode, msg.message))
			} else {
				d.stats.process(msg)
			}
		case <-d.done:
			nthreads -= 1
		case <-interrupt:
			fmt.Println("Received SIGINT, exiting...")
			close(d.stop)
		}
	}

	d.stats.Print()
}

type RequestCounter struct {
	count uint32
	mutex *sync.RWMutex
}

func (rc *RequestCounter) Add(i int) uint32 {
	rc.mutex.Lock()
	rc.count += uint32(i)
	rc.mutex.Unlock()
	return rc.count
}

func (rc *RequestCounter) Load() uint32 {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()
	return rc.count
}

type RunStats struct {
	counter *RequestCounter

	minTime time.Duration
	maxTime time.Duration

	total time.Duration
}

func (rs *RunStats) process(msg LogMessage) {
	if msg.ReqTime < rs.minTime {
		rs.minTime = msg.ReqTime
	}
	if rs.minTime == 0 {
		rs.minTime = msg.ReqTime
	}
	if msg.ReqTime > rs.maxTime {
		rs.maxTime = msg.ReqTime
	}

	rs.total += msg.ReqTime
}

func (rs *RunStats) Print() {
	fmt.Printf("Processed requests: %d\n", rs.counter.count)
	fmt.Printf("Max request time: %s\n", rs.maxTime)
	fmt.Printf("Min request time: %s\n", rs.minTime)

	if rs.counter.count > 0 {
		fmt.Printf("Average request time: %s\n", time.Duration(uint32(rs.total) / rs.counter.count))
	}
}

type ProxySettings struct {
	User, Password string
	URI            string
}

func ParseProxy(s string) (*url.URL, error) {
	return url.Parse(s)
}
