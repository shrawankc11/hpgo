package response

import (
	"bytes"
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
	writer         io.WriteCloser
	headers        *headers.Header
	headersWritten bool
}

func GetDefaultHeaders(cl int) *headers.Header {
	h := headers.NewHeaders()
	h.Set("Content-length", fmt.Sprintf("%d", cl))
	h.Set("Connection", "close")
	h.Set("Content-type", "text/plain")
	return h
}

func NewWriter(con net.Conn) *Writer {
	h := GetDefaultHeaders(0)

	return &Writer{
		writer:         con,
		headers:        h,
		headersWritten: false,
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

	if w.headersWritten {
		return fmt.Errorf("Cannot set header after they are already sent.")
	}

	var err error
	w.headers.Foreach(func(key, val string) {
		_, err = w.writer.Write([]byte(fmt.Appendf([]byte(""), "%s: %s\r\n", key, val)))
	})

	_, err = w.writer.Write([]byte("\r\n"))

	w.headersWritten = true

	if err != nil {
		return err
	}

	return nil
}

func (w *Writer) writeChuncked(data []byte) (int, error) {
	d := []byte("")
	d = fmt.Appendf(d, "%s", fmt.Sprintf("%x\r\n", len(data)))
	d = fmt.Appendf(d, "%s\r\n", data)
	w.writer.Write(d)
	return 0, nil
}

func (w *Writer) Write(data []byte) (int, error) {
	cl := w.headers.GetInt("Content-length")
	if w.headersWritten {
		if cl == 0 {
			reader := bytes.NewReader(data)
			buffer := make([]byte, 10)

			for reader.Len() > 0 {
				n, err := reader.Read(buffer)

				if err != nil {
					return 0, err
				}

				buffer = buffer[:n]
				w.writeChuncked(buffer)
			}

			n, err := w.writer.Write(fmt.Appendf([]byte(""), "0\r\n\r\n"))

			if err != nil {
				return 0, err
			}

			return n, nil
		}
	}

	if cl == 0 {
		cl = len(data)
		w.headers.Replace("Content-length", fmt.Sprintf("%d", cl))
	}

	w.WriteHeaders()

	n, err := w.writer.Write(data)

	if err != nil {
		return 0, err
	}
	fmt.Printf("n %d\n", n)
	return n, nil
}
