package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/httpfromtcp/internal/request"
	"github.com/httpfromtcp/internal/response"
	"github.com/httpfromtcp/internal/server"
)

const port uint16 = 8888

var TotalRequests uint = 1

func main() {
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		w.WriteStatusLine(response.StatusOk)
		if req.RequestLine.RequestTarget == "/get-chunked-res" {
			w.Header().Set("Transfer-encoding", "chunked")
			w.WriteHeaders()
		}
		buf := fmt.Appendf([]byte(""), "Hello, you are request number %d", TotalRequests)
		w.Write(buf)
		TotalRequests += 1
	})

	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
