package lib

import (
	"sync"
	"math/rand"
	"io/ioutil"
)

var letterBytes = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890")

func genRandomBytes(n uint) []byte {
	b := make([]byte, int(n))
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return b
}

type GrowableContent interface {
	// tries to grow payload of content according to content settings
	Grow() ([]byte, error)
}

type Payload struct {
	size          uint
	incValue      uint
	multiplyValue uint

	payload       []byte
	mutex         *sync.Mutex
}

func NewPayload(s, i, m uint) *Payload {
	return &Payload{s, i, m, nil, &sync.Mutex{}}
}

// Grow is the function that allows header payload grow according to it's settings. This function will be called after
// every successfull request for every emitter.
func (h *Payload) Grow() ([]byte, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if h.payload == nil {
		// first request
		h.payload = genRandomBytes(h.size)
		return h.payload, nil
	}

	if h.incValue > 0 {
		h.payload = append(h.payload, genRandomBytes(h.incValue)...)
		h.size += h.incValue
		return h.payload, nil
	}

	if h.multiplyValue > 1 {
		h.payload = append(h.payload, genRandomBytes(h.size * h.multiplyValue - uint(len(h.payload)))...)
		h.size *= h.multiplyValue
		return h.payload, nil
	}
	// header size doesn't grow
	return h.payload, nil
}

var _ GrowableContent = (*Payload)(nil)

const RequestHeaderName string = "Sample-Header"

type BodyFromFile struct {
	payload []byte
}

func NewBodyFromFile(f string) (*BodyFromFile, error) {
	payload, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	return &BodyFromFile{payload: payload}, nil
}

func (b *BodyFromFile) Grow() ([]byte, error) {
	return b.payload, nil
}

var _ GrowableContent = (*BodyFromFile)(nil)
