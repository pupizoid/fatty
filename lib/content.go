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

type Content struct {
	size          uint
	incValue      uint
	multiplyValue uint

	payload       []byte
	mutex         *sync.Mutex
}

func NewContent(s, i, m uint) *Content {
	return &Content{s, i, m, nil, &sync.Mutex{}}
}

// Grow is the function that allows content's payload grow according to it's settings.
func (h *Content) Grow() ([]byte, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if h.payload == nil {
		// first request
		h.payload = genRandomBytes(h.size)
		return h.payload, nil
	}
	// increment setting has higher priority
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
	// payload doesn't grow
	// todo: check for some settings overflow, like MaxHeaderSize or so on...
	return h.payload, nil
}

var _ GrowableContent = (*Content)(nil)

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
