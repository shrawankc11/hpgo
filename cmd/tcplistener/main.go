package main

import (
	"bytes"
	"io"
)

func ReadLine(read io.ReadCloser) <-chan string {
	out := make(chan string, 1)

	go func() {
		defer read.Close()
		defer close(out)

		str := ""
		for {
			data := make([]byte, 8)
			n, err := read.Read(data)

			if err != nil {
				break
			}

			data = data[:n]
			if i := bytes.IndexByte(data, '\n'); i != -1 {
				str += string(data[:i])
				data = data[i+1:]
				out <- str
				str = ""
			}

			str += string(data)
		}

		if len(str) != 0 {
			out <- str
		}

	}()

	return out
}
