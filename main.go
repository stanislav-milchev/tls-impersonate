package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	fhttp "github.com/Danny-Dasilva/fhttp"
	"github.com/Noooste/azuretls-client"
	urlverifier "github.com/davidmytton/url-verifier"
)

var verifier *urlverifier.Verifier

func init() {
	verifier = urlverifier.NewVerifier()
}

func main() {
	port := ":42069"
	log.Printf("Listening on localhost%s", port)
	fhttp.HandleFunc("/", HandleReq)
	fhttp.HandleFunc("/sleep", TimeoutChecker)
	err := fhttp.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalln("Error starting the HTTP server:", err)
	}
}

// TimeoutChecker is a helper endpoint to debug timeouts
func TimeoutChecker(w fhttp.ResponseWriter, r *fhttp.Request) {
	time.Sleep(time.Second * 45)
}

// HandleReq takes the incoming request, parses it, sends it towards the
func HandleReq(w fhttp.ResponseWriter, r *fhttp.Request) {
	stream := r.Header.Get("x-tls-stream") != ""

	session, req, err := NewRequest(r)
	if err != nil {
		log.Print(err)
		w.WriteHeader(fhttp.StatusBadRequest)
		return
	}

	defer session.Close()

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

	if err != nil {
		if strings.Contains(err.Error(), "timeout") {
			fmt.Print("timeout", err)
			w.WriteHeader(fhttp.StatusRequestTimeout)
			return
		} else {
			// TODO: EOF error encountered here at one point. Doesn't seem to happen now.
			// Potentially could be 'Connection' header issue
			fmt.Print("other error:", err)
			w.WriteHeader(res.StatusCode)
			w.Write(res.Body)
			return
		}
	}

	w.WriteHeader(res.StatusCode)

	// Either return buffered response or a stream
	if !stream {
		// Read the body and return buffered response
		if readBody, readErr := res.ReadBody(); readErr == nil {
			w.Write(readBody)
		} else {
			log.Printf("Error buffering response: %v", readErr)
		}
	} else {
		// Stream the response body
		_, err = io.Copy(w, res.RawBody)
		if err != nil {
			log.Printf("Error streaming response: %v", err)
		}

		// Close the response body
		res.RawBody.Close()
	}
}

func NewRequest(r *fhttp.Request) (*azuretls.Session, *azuretls.Request, error) {
	// Open and set-up session
	session := azuretls.NewSession()
	session.EnableLog()

	// Parse and validate request URL
	urlHeader := r.Header.Get("x-tls-url")
	res, err := verifier.Verify(urlHeader)
	if err != nil || res.IsURL == false {
		return nil, nil, errors.New("No valid request URL supplied via 'x-tls-url'; skipping request")
	}

	// Parse redirects
	disableRedirects := r.Header.Get("x-tls-disable-redirects") != ""

	// Parse timeout
	timeoutHeader := r.Header.Get("x-tls-timeout")
	t, err := strconv.Atoi(timeoutHeader)
	if err != nil || t <= 0 {
		// Probably dont log that on every request? Do it once and disable a flag or sth
		// log.Println("Invalid timeout value supplied, defaulting to 30s.")
		t = 30
	}
	timeout := time.Duration(t) * time.Second
	session.SetTimeout(timeout)

	// proxy := r.Header.Get("x-tls-proxy")

	req := &azuretls.Request{
		Method:           r.Method,
		Url:              urlHeader,
		DisableRedirects: disableRedirects,
		IgnoreBody:       true,
	}
	return session, req, nil
}
