package lib

import (
	"net/url"
	"github.com/valyala/fasthttp"
	"encoding/base64"
	"net/http"
	"errors"
	"fmt"
	"time"
)

// Base emitter interface
type Emitter interface {
	Start(stop, done chan struct{}, log chan EmitterEvent)
}
// Base emitter run options interface
type EmitterOptions interface {}

// Base emitter log message interface
type EmitterEvent interface {}

// Base statistic interface for emitter
type EmitterStats interface {}

type BaseEmitter struct {

}

// Load testing emitter

type LoadEmitter struct {
	client *fasthttp.HostClient
	proxy *url.URL

	options *LoadEmitterOptions
}

type LoadEmitterOptions struct {
	Urls chan string
	Ip string
	Port string
}

type LoadEmitterEvent struct {
	Code int
	RequestTime time.Duration
	RequestLength int
}

func NewLoadEmitter(options *LoadEmitterOptions, proxy *url.URL) Emitter {
	emitter := &LoadEmitter{}
	emitter.options = options
	emitter.proxy = proxy
	//emitter.client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxy)}}
	emitter.client = &fasthttp.HostClient{Addr: proxy.Host}
	return emitter
}
//
//func (e *LoadEmitter) Start(stop, done chan struct{}, log chan EmitterEvent) {
//
//	var req *http.Request
//	var resp *http.Response
//	var err error
//
//	for {
//		select {
//		case u := <-e.options.Urls:
//
//			req, err = http.NewRequest("GET", u, nil)
//
//			start := time.Now()
//			resp, err = e.client.Do(req)
//			if err != nil {
//				log <- errors.New(fmt.Sprintf("Error: %s", err))
//				//done <- struct{}{}
//				//return
//			}
//
//			ev := LoadEmitterEvent{
//				Code: resp.StatusCode,
//				RequestTime: time.Since(start),
//				RequestLength: resp.ContentLength,
//			}
//
//			//fmt.Printf("Event: %#v", ev)
//
//			log <- ev
//
//			select {
//			case _, ok := <-stop:
//				if !ok {
//					// channel closed so exit
//					done <- struct{}{}
//					return
//				}
//			default:
//			}
//
//			resp.Body.Close()
//		default:
//			// send message that this emitter has done all work
//			done <- struct{}{}
//			return
//		}
//	}
//}

func (e *LoadEmitter) Start(stop, done chan struct{}, log chan EmitterEvent) {

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	if e.proxy != nil && e.proxy.User != nil {
		req.Header.Set("Proxy-Authentication", base64.URLEncoding.EncodeToString([]byte(e.proxy.User.String())))
	}

	req.Header.SetMethod(http.MethodGet)

	for {
		select {
		case u := <-e.options.Urls:

			req.SetRequestURI(u)

			start := time.Now()
			err := e.client.Do(req, resp)
			if err != nil {
				log <- errors.New(fmt.Sprintf("Error: %s", err))
				//done <- struct{}{}
				//return
			}

			ev := LoadEmitterEvent{
				Code: resp.StatusCode(),
				RequestTime: time.Since(start),
				RequestLength: resp.Header.ContentLength(),
			}

			//fmt.Printf("Event: %#v", ev)

			log <- ev

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
		default:
			// send message that this emitter has done all work
			done <- struct{}{}
			return
		}
	}
}