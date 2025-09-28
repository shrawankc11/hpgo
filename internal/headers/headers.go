package headers

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

var SEPARATOR = "\r\n"
var SEPARATORN = SEPARATOR + SEPARATOR

type Headers map[string]string

type Header struct {
	headers Headers
}

func (h *Header) Foreach(cb func(key, val string)) {
	for k, v := range h.headers {
		cb(k, v)
	}
}

func (h *Header) Replace(key, val string) {
	h.headers[strings.ToLower(key)] = val
}

func (h *Header) Set(key, value string) {
	if v := h.Get(key); len(v) > 1 {
		newVal := fmt.Sprintf("%s, %s", v, value)
		h.headers[strings.ToLower(key)] = newVal
	} else {
		h.headers[strings.ToLower(key)] = value
	}
}

func (h *Header) Get(key string) string {
	return h.headers[strings.ToLower(key)]
}

func (h *Header) GetInt(key string) int {
	val := h.Get(key)
	intVal, _ := strconv.Atoi(val)
	return intVal
}

func (h *Header) Has(key string) bool {
	_, ok := h.headers[strings.ToLower(key)]
	return ok
}

func (h *Header) isTokenValid(value string) bool {
	for _, ch := range value {
		valid := false
		if ch > 'A' && ch < 'Z' || ch > 0 && ch < 9 || ch > 'a' && ch < 'z' {
			valid = true
		}
		switch ch {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			valid = true
		}

		if !valid {
			return valid
		}
	}

	return true
}

func (h Header) Parse(b []byte) (int, error) {
	//total read
	read := 0
	for {
		rnIdx := bytes.Index(b, []byte(SEPARATOR))

		//handle malformed error
		if rnIdx == -1 {
			return read, fmt.Errorf("malformed header")
		}

		//handle end of header
		if rnIdx == 0 {
			return read + len(SEPARATOR), nil
		}

		headerLine := b[:rnIdx]
		headerParts := bytes.SplitN(headerLine, []byte(":"), 2)
		fieldName := string(headerParts[0])

		if strings.HasSuffix(fieldName, " ") {
			return read, fmt.Errorf("malformed field name")
		}

		key := strings.ToLower(fieldName)
		val := strings.Trim(string(headerParts[1]), " ")
		h.Set(key, val)

		nextLine := rnIdx + len(SEPARATOR)
		read += nextLine
		b = b[nextLine:]
	}
}

func NewHeaders() *Header {
	headers := make(map[string]string)
	header := &Header{
		headers: headers,
	}

	return header
}
