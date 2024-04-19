package main

import (
	"fmt"
    fhttp "github.com/Danny-Dasilva/fhttp"
	"github.com/Danny-Dasilva/CycleTLS/cycletls"
	"io"
	"log"
)

func main() {
	port := ":42069"
	log.Printf("Listening on port %s", port)
	fhttp.HandleFunc("/", HandleReq)

	err := fhttp.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalln("Error starting the HTTP server:", err)
	}
}

// HandleReq takes the incoming request, parses it, sends it towards the
func HandleReq(w fhttp.ResponseWriter, r *fhttp.Request) {
	url := r.Header.Get("tls-request")
	method := r.Method

	req, err := fhttp.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	ja3 := "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-17513-18-16-10-23-13-65281-43-5-51-11-65037-27-35-45,29-23-24,0"
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 OPR/107.0.0.0"

	client := &fhttp.Client{
		Transport: cycletls.NewTransport(ja3, ua),
	}

	res, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer res.Body.Close()

	contentType := res.Header.Get("content-type")
	contentEncoding := res.Header.Get("content-encoding")

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Encoding", contentEncoding)

	body, err := io.ReadAll(res.Body)

	if err != nil {
		fmt.Println(err)
		return
	}

	w.Write(body)

}
