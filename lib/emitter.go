package lib

import (
	"net/http"
	"github.com/valyala/fasthttp"
	"time"
)

type Emitter struct {
	Destination string
	Method      string

	limit       uint32

	client      *fasthttp.Client

	header      GrowableContent
	body        GrowableContent
}

func NewEmitter(d, m string, l uint32, h, b GrowableContent) *Emitter {
	return &Emitter{Destination: d, Method: m, limit: l, client: &fasthttp.Client{}, header: h, body: b}
}

func (e *Emitter) Start(counter *RequestCounter, stop, done chan struct{}, log chan LogMessage) {

	var start time.Time

	req := fasthttp.AcquireRequest()
	req.SetRequestURI(e.Destination)
	req.Header.SetMethod(e.Method)

	resp := fasthttp.AcquireResponse()

	for counter.Load() < e.limit {
		// apply header if it persists in run configuration
		if e.header != nil {
			headerContent, err := e.header.Grow()
			if err != nil {
				// this can happen in case of overflow or reaching max header size...
				log <- LogMessage{err: true, message: err.Error()}
				done <- struct{}{}
				return
			}
			req.Header.SetBytesV(RequestHeaderName, headerContent)
		}
		// apply body to request if it persists in run configuration
		if e.body != nil && e.Method == http.MethodPost {
			bodyContent, err := e.body.Grow()
			if err != nil {
				// this can happen in case of overflow or reaching max header size...
				log <- LogMessage{err: true, message: err.Error()}
				done <- struct{}{}
				return
			}
			req.AppendBody(bodyContent)
		}

		counter.Add(1)
		start = time.Now()

		err := e.client.Do(req, resp)
		if err != nil {
			log <- LogMessage{err: true, message: err.Error()}
			done <- struct{}{}
			return
		}
		estimated := time.Since(start)

		switch code := resp.StatusCode(); code {
		case http.StatusFound, http.StatusOK:
			log <- LogMessage{ReqTime: estimated, ReqCode: code, err: false}
		default:
			log <- LogMessage{ReqTime: estimated, ReqCode: code, err: true, message: string(resp.Body())}
			done <- struct{}{}
			return
		}

		select {
		case _, ok := <-stop:
			if !ok {
				// channel closed so exit
				done <- struct{}{}
				return
			}
		default:
		}

		resp.Reset()
	}
	done <- struct{}{}
}

// todo: implement log message interface
type LogMessage struct {
	ReqTime time.Duration
	ReqCode int

	err     bool
	message string
}

