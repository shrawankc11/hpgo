package request

import (
	"bytes"
	"io"
	"strconv"

	"github.com/httpfromtcp/internal/headers"
)

type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

type Param struct {
	params map[string]string
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Header
	Body        []byte
	Complete    bool
	Params      *Param
}

type RequestState string

const (
	RequestStateRequestList RequestState = "request-line"
	RequestStateHeader      RequestState = "header"
	RequestStateBody        RequestState = "body"
)

var SEPARATOR = "\r\n"

func MakeRequest() *Request {
	params := make(map[string]string)

	rParams := &Param{
		params: params,
	}
	return &Request{
		Complete: false,
		Params:   rParams,
	}
}

func (p *Param) Get(key string) string {
	return p.params[key]
}

func ParseRequestLine(b []byte) (*RequestLine, int, error) {
	rnIdx := bytes.Index(b, []byte(SEPARATOR))
	rlByte := b[:rnIdx]

	parts := bytes.Split(rlByte, []byte(" "))

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(parts[2]),
	}

	return rl, rnIdx + len([]byte(SEPARATOR)), nil
}

func (r *Request) ParseBody(b []byte) (int, error) {
	return 0, nil
}

func (r *Request) ParseHeaders(b []byte) (int, error) {
	headers := headers.NewHeaders()
	currIdx, err := headers.Parse(b)

	if err != nil {
		return currIdx, err
	}

	r.Headers = headers
	return currIdx, nil
}

func getHeaderInt(h *headers.Header, key string, defaultValue int) int {
	valStr := h.Get(key)
	valInt, err := strconv.Atoi(valStr)

	if err != nil {
		return defaultValue
	}

	return valInt
}

func (r *Request) Parse(b []byte) (int, error) {
	var currIndex int
	rl, currIndex, err := ParseRequestLine(b)

	if err != nil {
		return currIndex, err
	}

	r.RequestLine = *rl

	b = b[currIndex:]
	currIndex, err = r.ParseHeaders(b)

	if err != nil {
		return currIndex, err
	}

	contentLength := getHeaderInt(r.Headers, "content-length", 0)

	if contentLength != 0 {
		r.Body = b[currIndex:]
	}

	r.Complete = true

	return currIndex, nil
}

func RequestFromReader(read io.ReadCloser) (*Request, error) {
	request := MakeRequest()

	buffer := make([]byte, 1024)
	bufLen := 0

	for !request.Complete {
		n, err := read.Read(buffer[bufLen:])

		if err != nil {
			return nil, err
		}

		bufLen += n

		idx, err := request.Parse(buffer)

		if err != nil {
			return nil, err
		}

		buffer = buffer[idx:bufLen]
		bufLen -= idx
	}

	return request, nil
}
