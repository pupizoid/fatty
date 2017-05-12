package lib

import (
	"net/http"
	"time"
	"net/url"
	"fmt"
	"encoding/base64"
	"github.com/valyala/fasthttp"
)

type Emitter struct {
	Destination *url.URL
	Method      string

	limit uint32

	client *fasthttp.Client

	header GrowableContent
	body   GrowableContent
	proxy  *url.URL
}

func NewEmitter(m string, l uint32, h, b GrowableContent, d, p *url.URL) *Emitter {
	client := &fasthttp.Client{}

	if p != nil {

	}

	return &Emitter{Destination: d, Method: m, limit: l, client: client, header: h, body: b, proxy: p}
}

func (e *Emitter) Start(counter *RequestCounter, proxy *url.URL, stop, done chan struct{}, log chan LogMessage) {

	var start time.Time

	req := fasthttp.AcquireRequest()

	if e.proxy != nil && e.proxy.User != nil {
		req.Header.Set("Proxy-Authentication", base64.URLEncoding.EncodeToString([]byte(e.proxy.User.String())))
	}

	req.Header.SetMethod(e.Method)

	// test

	req.Header.SetHost(e.Destination.Host)

	// test end

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
				// this can happen in case of overflow or reaching max body size...
				log <- LogMessage{err: true, message: err.Error()}
				done <- struct{}{}
				return
			}
			//req.AppendBody(bodyContent)
			req.SetBody(bodyContent)
		}

		counter.Add(1)
		start = time.Now()

		err := e.client.Do(req, resp)
		if err != nil {
			log <- LogMessage{err: true, message: fmt.Sprintf("Error processing request: %s", err.Error())}
			done <- struct{}{}
			return
		}
		time_spent := time.Since(start)

		// parse request status
		switch code := resp.StatusCode(); code {
		case http.StatusFound, http.StatusOK:
			log <- LogMessage{ReqTime: time_spent, ReqCode: code, err: false}
		default:
			log <- LogMessage{ReqTime: time_spent, ReqCode: code, err: true, message: string(resp.Body())}
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
			// can work further, so clear response
			resp.Reset()
		}
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
