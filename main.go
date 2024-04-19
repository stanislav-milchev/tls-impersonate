package main

import (
	"fmt"
	"io"
	"log"

	http "github.com/Carcraftz/fhttp"
)

func main() {
	port := ":42069"
	log.Printf("Listening on port %s", port)
	http.HandleFunc("/", HandleReq)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalln("Error starting the HTTP server:", err)
	}
}

// HandleReq takes the incoming request, parses it, sends it towards the
func HandleReq(w http.ResponseWriter, r *http.Request) {
	url := r.Header.Get("tls-request")
	method := r.Method
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	client := &http.Client{}
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
