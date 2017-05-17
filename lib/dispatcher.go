package lib

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"
	"net/url"
	"github.com/uber-go/atomic"
)

type Dispatcher struct {
	stats *LoadRunStats
	Proxy *url.URL

	Emitters []Emitter

	log chan EmitterEvent

	deadLine *time.Timer

	stop, done chan struct{}
}

func NewDispatcher(timeout int) *Dispatcher {
	//stats := &RunStats{counter: &RequestCounter{mutex: &sync.RWMutex{}}}
	stats := &LoadRunStats{
		counter: atomic.NewInt32(0),
		statusCodes: make(map[int]int),
		startTime: time.Now(),
	}

	timer := time.Timer{}
	if timeout != 0 {
		timer = *time.NewTimer(time.Second * time.Duration(timeout))
	}

	done := make(chan struct{})
	stop := make(chan struct{})

	return &Dispatcher{stats: stats, done: done, stop: stop, deadLine: &timer}
}

func (d *Dispatcher) Run() {

	nthreads := len(d.Emitters)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	d.log = make(chan EmitterEvent, 10000)

	for _, e := range d.Emitters {
		go e.Start(d.stop, d.done, d.log)
	}

	for nthreads > 0 {
		select {
		case event := <-d.log:
			switch msg := event.(type) {
			case LoadEmitterEvent:
				if _, ok := d.stats.statusCodes[msg.Code]; ok {
					d.stats.statusCodes[msg.Code]++
				} else {
					d.stats.statusCodes[msg.Code] = 1
				}
				d.stats.totalTime += msg.RequestTime
				if d.stats.minRequestTime > msg.RequestTime || d.stats.minRequestTime == 0 {
					d.stats.minRequestTime = msg.RequestTime
				}
				if d.stats.maxRequestTime < msg.RequestTime {
					d.stats.maxRequestTime = msg.RequestTime
				}
				d.stats.totalBytes += msg.RequestLength
				d.stats.counter.Add(1)
			case error:
				d.stats.errorCounter++
			default:
				fmt.Printf("Unknown event: %#v", msg)
			}
		case <-d.done:
			nthreads -= 1
		case <-interrupt:
			fmt.Println("Received SIGINT, exiting...")
			close(d.stop)
		case <-d.deadLine.C:
			fmt.Println("Testing ended by time")
			close(d.stop)
		default:
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

type LoadRunStats struct {
	counter *atomic.Int32
	errorCounter int
	statusCodes map[int]int
	totalTime time.Duration
	totalBytes int

	maxRequestTime time.Duration
	minRequestTime time.Duration

	startTime time.Time
}

func (rs *LoadRunStats) Print() {
	fmt.Printf("Processed requests: %d\n", rs.counter.Load())
	fmt.Println("Status code information:")
	for code, count := range rs.statusCodes {
		fmt.Printf("%d: %d\n", code, count)
	}
	fmt.Printf("Connection errors: %d\n", rs.errorCounter)
	// count cumulative values
	bandwidth := float64(rs.totalBytes) / rs.totalTime.Seconds()
	fmt.Printf("Max request time: %s\n", rs.maxRequestTime)
	fmt.Printf("Min request time: %s\n", rs.minRequestTime)
	fmt.Printf("Overal bandwidth: %f bytes/sec\n", bandwidth)
	fmt.Printf("Time: %s\n", time.Since(rs.startTime))
}

type ProxySettings struct {
	User, Password string
	URI            string
}

func ParseProxy(s string) (*url.URL, error) {
	return url.Parse(s)
}
