// Package allows you to test any server or proxy for it's max header length
// and request size
package main

import (
	"net/http"
	"math/rand"
	"fmt"
	"encoding/base64"
	"bytes"
	"io/ioutil"
	"github.com/mkideal/cli"
	"github.com/c2h5oh/datasize"
	"bitbucket.org/taruti/mimemagic"
	clix "github.com/mkideal/cli/ext"
	"os"
	"log"
	"time"
)

type (
	TestingEntity interface {
		Next()
		SetBody(*http.Request)
		SetHeader(*http.Request)
		GetCurrentSize() int
		GetPrevSize() int
		CurrentInfo() string
	}

	GrovingHeader struct {
		strategy string
		ratio int
		size  int
	}

	GrovingBody struct {
		strategy string
		ratio int
		size  int
	}

	FileBody struct {
		file clix.File
		size int
		mime string
	}

	TestEntity struct {
		strategy string
		ratio int
		size  int
	}

	TestingOptions struct {
		cli.Helper
		Multiplier   int    `cli:"m,multi" usage:"sets payload multiplier" dft:"2"`
		Strategy     string `cli:"*s,strategy" usage:"strategy of payload groving (linear, exponential)"`
		Entites      []TestingEntity
		Method       string `cli:"*x,method" usage:"request method to use"`
		Dest         string `cli:"*d,dest" usage:"destination of requests"`
		Proxy        string `cli:"proxy" usage:"addr of proxy server to use"`
		ProxyUsr     string `cli:"proxy-user" usage:"username for proxy-authenticate"`
		ProxyPwd     string `cli:"proxy-pwd" usage:"password of proxy user"`
		RequestLimit int    `cli:"l,limit" usage:"limit of requests"`
		HeaderSize   readableSize `cli:"header" usage:"if specified adds header with given initial value size" dft:"0" parser:"datasize.UnmarchalText"`
		BodySize     readableSize `cli:"b,body" usage:"if specified adds random payload with given initial size" dft:"0" parser:"datasize.UnmarchalText"`
		FileContent  clix.File `cli:"p,payload-file" usage:"overrides random payload of body by file contents"`
		Verbose      bool   `cli:"verbose" usage:"verbose flag"`
	}

	ProxySettings struct {
		URL, User, Pass string
	}

	readableSize struct {
		size datasize.ByteSize
	}
)

func (rs *readableSize) Decode(s string) error {
	return rs.size.UnmarshalText([]byte(s))
}

func (rs readableSize) NotNull() bool {
	return rs.size != 0x0
}

func getRandomValue(size int) []byte {
	letterRunes := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]byte, size)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return b
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
	switch gh.strategy {
	case "linear":
		payload = gh.size - gh.ratio
	case "exponential":
		payload = gh.size / gh.ratio
	}
	return payload
}

func (gh *GrovingHeader) Next() {
	switch gh.strategy {
	case "linear":
		gh.size += gh.ratio
	case "exponential":
		gh.size = gh.size * gh.ratio
	}
}

func (gh *GrovingHeader) CurrentInfo() string {
	return fmt.Sprintf("Header testing size - %d", gh.GetCurrentSize())
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
	switch gb.strategy {
	case "linear":
		payload = gb.size - gb.ratio
	case "exponential":
		payload = gb.size / gb.ratio
	}
	return payload
}

func (gb *GrovingBody) Next() {
	switch gb.strategy {
	case "linear":
		gb.size += gb.ratio
	case "exponential":
		gb.size = gb.size * gb.ratio
	}
}

func (gb *GrovingBody) CurrentInfo() string {
	return fmt.Sprintf("Body testing size - %d", gb.GetCurrentSize())
}

func (fb *FileBody) SetBody(req *http.Request) {
	req.Body = ioutil.NopCloser(bytes.NewReader(fb.file.Data()))
}

func (fb *FileBody) SetHeader(req *http.Request) {
	if fb.mime != "" {
		req.Header.Add("Content-Type", fb.mime)
	}
}

func (fb *FileBody) GetCurrentSize() int { return fb.size }

func (fb *FileBody) GetPrevSize() int { return fb.size }

func (fb *FileBody) Next() {}

func (fb *FileBody) CurrentInfo() string {
	return fmt.Sprintf("Body testing size - %d, mime type - %s", fb.GetCurrentSize(), fb.mime)
}

func processRequest(c *http.Client, options *TestingOptions) {
	req, err := http.NewRequest(options.Method, options.Dest, nil)
	if err != nil {
		log.Fatalf("Error creating request to the server: %v\nProgram terminated\n", err)
	}
	if options.ProxyUsr != "" && options.ProxyPwd != "" {
		req.Header.Add("Proxy-Authorization", fmt.Sprintf(
			"Basic %s", base64.StdEncoding.EncodeToString([]byte(options.ProxyUsr + ":" + options.ProxyPwd))))
	}
	// apply payload
	for _, entity := range options.Entites {
		entity.SetHeader(req)
		entity.SetBody(req)
		//if options.Verbose {
		//	log.Printf("setting request data by %#v\n", entity)
		//}
	}
	// checks
	for _, entity := range options.Entites {
		switch entity.(type) {
		case *GrovingBody:
			if options.Method == "HEAD" || options.Method == "GET" {
				log.Fatal("Testing body size not allowed for this method")
			}
		}
	}

	//
	resp, err := c.Do(req)
	if err != nil {
		log.Fatalf("Error creating request to the server: %v\nProgram terminated\n", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("Received %d request status code, terminating...\n", resp.StatusCode)
		for _, entity := range options.Entites {
			log.Println(entity.CurrentInfo())
		}
		//log.Printf("Testing size - %d\n", options.strategy.GetCurrentSize())
		//log.Printf("Last successful testing size - %d\n", options.strategy.GetPrevSize())
		os.Exit(0)
	} else if options.Verbose {
		//log.Printf("Successful request, testing size - %d", options.strategy.GetCurrentSize())
		log.Print("Successful request. ")
		for _, entity := range options.Entites {
			log.Print(entity.CurrentInfo())
		}
		log.Print("\n")
	}
	resp.Body.Close()
	for _, entity := range options.Entites {
		entity.Next()
	}
}

func main() {

	cli.Run(new(TestingOptions), func(ctx *cli.Context) error {
		options := ctx.Argv().(*TestingOptions)
		// check strategy
		if !(options.Strategy == "linear" || options.Strategy == "exponential") {
			return fmt.Errorf("%s: only 'linear' or 'exponential' supported.", ctx.Color().Red("Err"))
		}

		if options.BodySize.NotNull() && options.FileContent.String() == "" {
			options.Entites = append(options.Entites, &GrovingBody{strategy: options.Strategy,
				ratio: options.Multiplier, size: int(options.BodySize.size.Bytes())})
		} else if options.FileContent.String() != "" {
			options.Verbose = true
			options.RequestLimit = 1
			fileBody := &FileBody{
				file: options.FileContent,
				mime: mimemagic.Match(options.FileContent.String(), options.FileContent.Data()),
				size: len(options.FileContent.Data()),
			}
			options.Entites = append(options.Entites, fileBody)
		}

		if options.HeaderSize.NotNull() {
			options.Entites = append(options.Entites, &GrovingHeader{strategy: options.Strategy,
				ratio: options.Multiplier, size: int(options.HeaderSize.size.Bytes())})
		}

		if len(options.Entites) == 0 {
			return fmt.Errorf("%s: no testing entity specified, please set body or header size", ctx.Color().Red("Err"))
		}

		//fmt.Printf("%#v\n", options)

		client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment}, Timeout: time.Second * 10}
		if options.Proxy != "" {
			os.Setenv("HTTP_PROXY", options.Proxy)
		}

		reqCount := 0
		for {
			processRequest(client, options)
			reqCount++
			if options.RequestLimit != 0 && options.RequestLimit <= reqCount {
				log.Printf("Successfuly made %d requests", reqCount)
				os.Exit(0)
			}
		}

		return nil
	})





	//flag.StringVar(&options.Dest, "dest", "", "Destination of a testing server")
	//flag.StringVar(&options.Method, "method", "", "Request method")
	//flag.StringVar(&options.Proxy.URL, "proxy", "", "Destination of a proxy server")
	//flag.StringVar(&options.Proxy.User, "proxy-user", "", "Proxy username")
	//flag.StringVar(&options.Proxy.Pass, "proxy-pass", "", "Proxy password")
	//flag.IntVar(&options.Ratio, "r", 2, "Testing value increase ratio")
	//flag.IntVar(&options.InitialSize, "i", 1, "Testing value initial size")
	//flag.IntVar(&options.MaxRequests, "m", 0, "Max number of requests to perform [0 = limitless]")
	//flag.BoolVar(&options.Verbose, "v", false, "Verbosive output")
	//
	//var (
	//	testHeader, testBody int
	//	strategy string
	//)
	//flag.IntVar(&testHeader, "h", false, "Test header size")
	//flag.IntVar(&testBody, "b", false, "Test body size")
	//flag.StringVar(&strategy, "strategy", "", "Testing value groving strategy [linear,expo]")
	//
	//flag.Parse()
	//
	//if strategy != "linear" && strategy != "expo" {
	//	log.Fatal("Unsupported testing strategy")
	//}
	//if testBody && !testHeader {
	//	log.Println("Testing request Body with groving size:")
	//	options.Strategy = &GrovingBody{
	//		ratio: options.Ratio,
	//		size: options.InitialSize,
	//		strategy: strategy,
	//	}
	//} else if !testBody && testHeader {
	//	log.Println("Testing request Header with groving size:")
	//	options.Strategy = &GrovingHeader{
	//		ratio: options.Ratio,
	//		size: options.InitialSize,
	//		strategy: strategy,
	//	}
	//} else if testBody && testHeader {
	//	log.Println("Testing request Body and Header with groving size:")
	//	options.Strategy = &MixedGrovingTest{
	//		ratio: options.Ratio,
	//		size: options.InitialSize,
	//		strategy: strategy,
	//	}
	//} else {
	//	log.Fatal("Specify entity to test (header,body or both)")
	//}
	//
	//client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment}}
	//if options.Proxy.URL != "" {
	//	os.Setenv("HTTP_PROXY", options.Proxy.URL)
	//}
	//
	//reqCount := 0
	//for {
	//	processRequest(client, options)
	//	reqCount++
	//	if options.MaxRequests != 0 && options.MaxRequests < reqCount {
	//		log.Printf("Successfuly made %d requests", reqCount)
	//		os.Exit(0)
	//	}
	//}
}
