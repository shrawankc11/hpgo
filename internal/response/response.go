package response

import (
	"fmt"
	"io"
	"net"

	"github.com/httpfromtcp/internal/headers"
)

type StatusCode uint16

const (
	StatusOk                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type Writer struct {
	writer       io.WriteCloser
	headers      *headers.Header
	wroteHeaders bool
}

func GetDefaultHeaders(cl int) *headers.Header {
	h := headers.NewHeaders()
	h.Set("Content-length", fmt.Sprintf("%d", cl))
	h.Set("Connection", "close")
	h.Set("Content-type", "text/plain")
	return h
}

func NewWriter(con net.Conn) *Writer {
	h := headers.NewHeaders()
	h.Set("Content-length", fmt.Sprintf("%d", 0))
	h.Set("Connection", "close")
	h.Set("Content-type", "text/plain")

	return &Writer{
		writer:       con,
		headers:      h,
		wroteHeaders: false,
	}
}

func (w *Writer) WriteStatusLine(status StatusCode) error {
	responseMap := map[StatusCode]string{
		200: "OK",
		400: "Bad Request",
		500: "Internal Server Error",
	}

	rp := responseMap[status]

	if len(rp) == 0 {
		return fmt.Errorf("invalid reason phrase")
	}

	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", status, responseMap[status])
	_, err := w.writer.Write([]byte(statusLine))

	if err != nil {
		return err
	}

	return nil
}

func (w *Writer) Header() *headers.Header {
	return w.headers
}

func (w *Writer) WriteHeaders() error {
	var err error
	w.headers.Foreach(func(key, val string) {
		_, err = w.writer.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, val)))
	})

	_, err = w.writer.Write([]byte("\r\n"))

	w.wroteHeaders = true

	if err != nil {
		return err
	}

	return nil
}

func (w *Writer) Write(data []byte) (int, error) {
	cl := len(data)

	w.headers.Set("Content-length", fmt.Sprintf("%d", cl))

	if !w.wroteHeaders {
		w.WriteHeaders()
	}

	n, err := w.writer.Write(data)
	if err != nil {
		return 0, err
	}
	fmt.Printf("n %d\n", n)
	return n, nil
}
