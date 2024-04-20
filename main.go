package main

import (
	"errors"
	"log"
	"strconv"
	"time"

	fhttp "github.com/Danny-Dasilva/fhttp"
	"github.com/Noooste/azuretls-client"
	urlverifier "github.com/davidmytton/url-verifier"
)

var session *azuretls.Session
var verifier *urlverifier.Verifier

func init() {
	session = azuretls.NewSession()
	verifier = urlverifier.NewVerifier()
}

func main() {
	port := ":42069"
	log.Printf("Listening on localhost%s", port)
	fhttp.HandleFunc("/", HandleReq)
	err := fhttp.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalln("Error starting the HTTP server:", err)
	}
}

// HandleReq takes the incoming request, parses it, sends it towards the
func HandleReq(w fhttp.ResponseWriter, r *fhttp.Request) {
	req, err := NewRequest(r)
	if err != nil {
		log.Print(err)
		w.WriteHeader(fhttp.StatusBadRequest)
		return
	}
	/*
		ja3 := "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-17513-18-16-10-23-13-65281-43-5-51-11-65037-27-35-45,29-23-24,0"
		ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 OPR/107.0.0.0"
	*/

	// TODO: update default browser headers and put them in a separate package
	session.OrderedHeaders = azuretls.OrderedHeaders{
		{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		{"accept-encoding", "gzip, deflate, br"},
		{"accept-language", "zh-HK,zh-TW;q=0.9,zh;q=0.8"},
		{"cache-control", "max-age=0"},
		{"sec-ch-ua", `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`},
		{"sec-ch-ua-mobile", "?0"},
		{"sec-ch-ua-platform", `"Windows"`},
		{"sec-fetch-site", "none"},
		{"sec-fetch-mode", "navigate"},
		{"sec-fetch-user", "?1"},
		{"sec-fetch-dest", "document"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"},
		{"upgrade-insecure-requests", "1"},
	}

	res, err := session.Do(req)

	if err != nil && err.Error() == "timeout" {
		w.WriteHeader(fhttp.StatusRequestTimeout)
		return
	} else if err != nil {
		// TODO: EOF error encountered here at one point. Doesn't seem to happen now.
		// Potentially could be 'Connection' header issue
		w.WriteHeader(res.StatusCode)
		w.Write(res.Body)
		return
	}

	contentType := res.Header.Get("content-type")
	// contentEncoding := res.Header.Get("content-encoding")

	w.Header().Set("Content-Type", contentType)
	// w.Header().Set("Content-Encoding", contentEncoding)

	w.WriteHeader(res.StatusCode)
	w.Write(res.Body)
}

func NewRequest(r *fhttp.Request) (*azuretls.Request, error) {
	// Parse and validate request URL
	url := r.Header.Get("x-tls-url")
	res, err := verifier.Verify(url)
	if err != nil || res.IsURL == false {
		return nil, errors.New("No valid request URL supplied via 'x-tls-url'; skipping request")
	}

	timeout := r.Header.Get("x-tls-timeout")
	// proxy := r.Header.Get("x-tls-proxy")
	// redirects := r.Header.Get("x-tls-redirects")
	method := r.Method

	// Parse timeout
	t, err := strconv.Atoi(timeout)
	if err != nil {
		// Probably dont log that on every request? Do it once and disable a flag or sth
		log.Println("Invalid timeout value supplied, defaulting to 30s.")
		t = 30
	}

	req := azuretls.Request{
		Method:           method,
		Url:              url,
		TimeOut:          time.Duration(t) * time.Second,
		DisableRedirects: true,
	}
	return &req, nil
}
