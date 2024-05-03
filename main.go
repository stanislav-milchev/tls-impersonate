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
	"github.com/stanislav-milchev/tls-impersonator/browser"
)

var verifier *urlverifier.Verifier

func init() {
	verifier = urlverifier.NewVerifier()
}

func main() {
	port := ":42069"
	log.Printf("Listening on localhost%s", port)
	fhttp.HandleFunc("/", HandleReq)
	// dev testing endpoints
	fhttp.HandleFunc("/sleep", TimeoutChecker)
	fhttp.HandleFunc("/headers", handleHeaderYoink)

	err := fhttp.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalln("Error starting the HTTP server:", err)
	}
}

// handleHeaderYoink is a helper endpoint to get the header values of the current request
func handleHeaderYoink(_ fhttp.ResponseWriter, r *fhttp.Request) {
	for header, value := range r.Header {
		fmt.Printf("{\"%s\", \"%s\"}\n", header, value[0])
	}
}

// TimeoutChecker is a helper endpoint to debug timeouts
func TimeoutChecker(w fhttp.ResponseWriter, r *fhttp.Request) {
	time.Sleep(time.Second * 45)
}

// HandleReq takes the incoming request, parses it, sends it towards the target host
func HandleReq(w fhttp.ResponseWriter, r *fhttp.Request) {
	stream := r.Header.Get("x-tls-stream") != ""

	session, req, err := NewRequest(r)
	if err != nil {
		log.Print(err)
		w.WriteHeader(fhttp.StatusBadRequest)
		return
	}

	defer session.Close()

	session.OrderedHeaders = browser.Chrome124
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

	// Forward the headers received
	w.WriteHeader(res.StatusCode)
	for h, v := range res.Header {
		// Response we get is already decoded so this header will only cause issues with the
		// client used for the request
		if "content-encoding" == strings.ToLower(h) {
			continue
		}
		if len(v) > 0 {
			w.Header().Set(h, v[0])
		} else {
			fmt.Printf("Skipping \"%s\" header with invalid value", h)
			continue
		}
	}

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

	// Parse proxy
	proxy := r.Header.Get("x-tls-proxy")
	session.SetProxy(proxy)

	req := &azuretls.Request{
		Method:           r.Method,
		Url:              urlHeader,
		DisableRedirects: disableRedirects,
		IgnoreBody:       true,
	}
	return session, req, nil
}
