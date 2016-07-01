// Package allows you to test any server or proxy for it's max header length
// and request size
package main

import (
	"net/http"
	"math/rand"
	"log"
	"fmt"
	"encoding/base64"
	"os"
	"bytes"
	"io/ioutil"
	"flag"
)

type (
	Strategy interface {
		Next()
		SetBody(*http.Request)
		SetHeader(*http.Request)
		GetCurrentSize() int
		GetPrevSize() int
	}

	GrovingHeader struct {
		_type string
		ratio int
		size  int
	}

	GrovingBody struct {
		_type string
		ratio int
		size  int
	}
	
	MixedGrovingTest struct {
		_type string
		ratio int
		size  int
	}
	
	TestEntity struct {
		_type string
		ratio int
		size  int
	}

	TestingOptions struct {
		Ratio       int
		Strategy    Strategy
		Method      string
		Dest        string
		Proxy       ProxySettings
		MaxRequests int
		InitialSize int
		Verbose     bool
	}

	ProxySettings struct {
		URL, User, Pass string
	}
)

func (ps *ProxySettings) getAuthField() string {
	return base64.StdEncoding.EncodeToString([]byte(ps.User + ":" + ps.Pass))
}

func getRandomValue(size int) []byte {
	letterRunes := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]byte, size)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return b
}

func NewGrovingHeader(start_size, ratio int, _type string) *GrovingHeader {
	return &GrovingHeader{ratio: ratio, size: start_size, _type: _type}
}

// Body is nil for groving header strategy
func (gh *GrovingHeader) SetBody(req *http.Request) {}

func (gh *GrovingHeader) SetHeader(req *http.Request) {
	req.Header.Set("X-Groving-Header", string(getRandomValue(gh.size)))
}

func (gh *GrovingHeader) GetCurrentSize() int {
	return gh.size
}

func (gh *GrovingHeader) GetPrevSize() int {
	var payload int 
	switch gh._type {
	case "linear":
	 	payload = gh.size - gh.ratio
	case "expo":
		payload = gh.size / gh.ratio
	}
	return payload
}

func (gh *GrovingHeader) Next() {
	switch gh._type {
	case "linear":
		gh.size += gh.ratio
	case "expo":
		gh.size = gh.size * gh.ratio
	}
}

func NewGrovingBody(start_size, ratio int, _type string) *GrovingBody {
	return &GrovingBody{ratio: ratio, size: start_size, _type: _type}
}

func (gb *GrovingBody) SetBody(req *http.Request) {
	req.Body = ioutil.NopCloser(bytes.NewReader(getRandomValue(gb.size)))
}

func (gb *GrovingBody) SetHeader(req *http.Request) {}

func (gb *GrovingBody) GetCurrentSize() int {
	return gb.size
}

func (gb *GrovingBody) GetPrevSize() int {
	var payload int 
	switch gb._type {
	case "linear":
	 	payload = gb.size - gb.ratio
	case "expo":
		payload = gb.size / gb.ratio
	}
	return payload
}

func (gb *GrovingBody) Next() {
	switch gb._type {
	case "linear":
		gb.size += gb.ratio
	case "expo":
		gb.size = gb.size * gb.ratio
	}
}

func NewMixedGrovingTest(ratio, start_size int, _type string) *MixedGrovingTest {
	return &MixedGrovingTest{ratio: ratio, size: start_size, _type: _type}
}

func (mg *MixedGrovingTest) SetBody(req *http.Request) {
	req.Body = ioutil.NopCloser(bytes.NewReader(getRandomValue(mg.size)))
}

func (mg *MixedGrovingTest) SetHeader(req *http.Request) {
	req.Header.Set("X-Groving-Header", string(getRandomValue(mg.size)))
}

func (mg *MixedGrovingTest) GetCurrentSize() int {
	return mg.size
}

func (mg *MixedGrovingTest) GetPrevSize() int {
	var payload int 
	switch mg._type {
	case "linear":
	 	payload = mg.size - mg.ratio
	case "expo":
		payload = mg.size / mg.ratio
	}
	return payload
}

func (mg *MixedGrovingTest) Next() {
	switch mg._type {
	case "linear":
		mg.size += mg.ratio
	case "expo":
		mg.size = mg.size * mg.ratio
	}
}

func processRequest(c *http.Client, options *TestingOptions) {
	req, err := http.NewRequest(options.Method, options.Dest, nil)
	if err != nil {
		log.Fatalf("Error creating request to the server: %v\nProgram terminated\n", err)
	}
	if options.Proxy.User != "" && options.Proxy.Pass != "" {
		req.Header.Add("Proxy-Authorization", fmt.Sprintf("Basic %s", options.Proxy.getAuthField()))
	}
	// apply payload
	options.Strategy.SetHeader(req)
	options.Strategy.SetBody(req)
	// checks
	
	//
	resp, err := c.Do(req)
	if err != nil {
		log.Fatalf("Error creating request to the server: %v\nProgram terminated\n", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("Received %d request status code, terminating...\n", resp.StatusCode)
		log.Printf("Testing size - %d\n", options.Strategy.GetCurrentSize())
		log.Printf("Last successful testing size - %d\n", options.Strategy.GetPrevSize())
		os.Exit(0)
	} else if options.Verbose {
		log.Printf("Successful request, testing size - %d", options.Strategy.GetCurrentSize())
	}
	resp.Body.Close()
	options.Strategy.Next()
}

func main() {
	options := new(TestingOptions)
	flag.StringVar(&options.Dest, "dest", "", "Destination of a testing server")
	flag.StringVar(&options.Method, "method", "", "Request method")
	flag.StringVar(&options.Proxy.URL, "proxy", "", "Destination of a proxy server")
	flag.StringVar(&options.Proxy.User, "proxy-user", "", "Proxy username")
	flag.StringVar(&options.Proxy.Pass, "proxy-pass", "", "Proxy password")
	flag.IntVar(&options.Ratio, "r", 2, "Testing value increase ratio")
	flag.IntVar(&options.InitialSize, "i", 1, "Testing value initial size")
	flag.IntVar(&options.MaxRequests, "m", 0, "Max number of requests to perform [0 = limitless]")
	flag.BoolVar(&options.Verbose, "v", false, "Verbosive output")

	var (
		testHeader, testBody bool
		strategy string
	)
	flag.BoolVar(&testHeader, "h", false, "Test header size")
	flag.BoolVar(&testBody, "b", false, "Test body size")
	flag.StringVar(&strategy, "strategy", "", "Testing value groving strategy [linear,expo]")

	flag.Parse()

	if strategy != "linear" && strategy != "expo" {
		log.Fatal("Unsupported testing strategy")
	}
	if testBody && !testHeader {
		log.Println("Testing request Body with groving size:")
		options.Strategy = &GrovingBody{
			ratio: options.Ratio,
			size: options.InitialSize,
			_type: strategy,
		}
	} else if !testBody && testHeader {
		log.Println("Testing request Header with groving size:")
		options.Strategy = &GrovingHeader{
			ratio: options.Ratio,
			size: options.InitialSize,
			_type: strategy,
		}
	} else if testBody && testHeader {
		log.Println("Testing request Body and Header with groving size:")
		options.Strategy = &MixedGrovingTest{
			ratio: options.Ratio,
			size: options.InitialSize,
			_type: strategy,
		}
	} else {
		log.Fatal("Specify entity to test (header,body or both)")
	}

	client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment}}
	if options.Proxy.URL != "" {
		os.Setenv("HTTP_PROXY", options.Proxy.URL)
	}

	reqCount := 0
	for {
		processRequest(client, options)
		reqCount++
		if options.MaxRequests != 0 && options.MaxRequests < reqCount {
			log.Printf("Successfuly made %d requests", reqCount)
			os.Exit(0)
		}
	}
}
