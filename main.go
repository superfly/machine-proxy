package main

import (
	"flag"
	"io"
	"log"
	"net/http"
)

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

type proxy struct {
}

func (p *proxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {

	//http: Request.RequestURI can't be set in client requests.
	//http://golang.org/src/pkg/net/http/client.go
	log.Println(req.RemoteAddr, " ", req.Method, " ", req.URL)

	req.RequestURI = ""
	req.URL.Host = "localhost:10201"
	req.URL.Scheme = "http"

	log.Println(req.RemoteAddr, " ", req.Method, " ", req.URL)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(wr, "Server Error", http.StatusInternalServerError)
		log.Fatal("ServeHTTP:", err)
	}
	defer resp.Body.Close()

	log.Println(req.RemoteAddr, " ", resp.Status)

	copyHeader(wr.Header(), resp.Header)
	wr.WriteHeader(resp.StatusCode)
	io.Copy(wr, resp.Body)
}

func main() {
	var addr = flag.String("addr", "127.0.0.1:8080", "The addr of the application.")
	flag.Parse()

	handler := &proxy{}

	log.Println("Starting proxy server on", *addr)
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
