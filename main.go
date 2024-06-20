package main

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	fhttp "github.com/Noooste/fhttp"
	"github.com/Noooste/azuretls-client"
	"github.com/stanislav-milchev/tls-impersonator/browser"
)

var (
	serverPort          = getEnv("TLS_PORT", "8082")
	urlHeaderName       = getEnv("TLS_URL", "x-tls-url")
	proxyHeaderName     = getEnv("TLS_PROXY", "x-tls-proxy")
	bufferingHeaderName = getEnv("TLS_BUFFER", "x-tls-buffer")
	redirectHeaderName  = getEnv("TLS_REDIRECT", "x-tls-allowredirect")
	timeoutHeaderName   = getEnv("TLS_TIMEOUT", "x-tls-timeout")
)

func main() {
    port := fmt.Sprintf(":%s", serverPort)
	log.Printf("Listening on localhost%s", port)
	fhttp.HandleFunc("/", HandleReq)
	fhttp.HandleFunc("/isalive", HandleIsAlive)

	err := fhttp.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalln("Error starting the HTTP server:", err)
	}
}

func HandleIsAlive(w fhttp.ResponseWriter, r *fhttp.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(fhttp.StatusOK)
	fmt.Fprintf(w, `{"isalive":true}`)
}

// HandleReq takes the incoming request, parses it, sends it towards the target host
func HandleReq(w fhttp.ResponseWriter, r *fhttp.Request) {
	session, req, err := NewRequest(r)
	if err != nil {
		log.Print(err)
		w.WriteHeader(fhttp.StatusBadRequest)
		return
	}

	defer session.Close()

	SetHeaders(session, r.Header)
	SetCookies(req.Url, session, r.Cookies())

	res, err := session.Do(req)

	if err != nil {
		if strings.Contains(err.Error(), "timeout") {
			fmt.Print("timeout\n", err)
			w.WriteHeader(fhttp.StatusRequestTimeout)
			return
		} else {
			fmt.Print("other error:\n", err)
			w.WriteHeader(fhttp.StatusInternalServerError)
			return
		}
	}

	// Forward the headers received
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

	var buffering bool
	switch b := r.Header.Get(bufferingHeaderName); b {
	case "true", "True", "1":
		buffering = true
	default:
		buffering = false

	}

	w.WriteHeader(res.StatusCode)
	// Either return buffered response or a stream
	if buffering {
		if readBody, readErr := res.ReadBody(); readErr == nil {
			w.Write(readBody)
		} else {
			log.Printf("Error buffering response: %v", readErr)
		}
	} else {
		_, err = io.Copy(w, res.RawBody)
		if err != nil {
			log.Printf("Error streaming response: %v", err)
		}

		res.RawBody.Close()
	}
}

// NewRequest opens a new azuretls session and a request, and sets it up with url,
// proxy, headers, cookies, redirects and timeouts
func NewRequest(r *fhttp.Request) (*azuretls.Session, *azuretls.Request, error) {
	// Open and set-up session
	session := azuretls.NewSession()
	session.EnableLog()

	// Parse URL
	urlHeader := r.Header.Get(urlHeaderName)

	if urlHeader == "" {
		return nil, nil, fmt.Errorf(
			"no valid request URL supplied via '%s'; skipping request", urlHeaderName,
		)
	}

	// Parse redirects
	var allowRedirects bool
	switch rH := r.Header.Get(redirectHeaderName); rH {
	case "true", "True", "1":
		allowRedirects = true
	default:
		allowRedirects = false

	}

	// Parse timeout
	timeoutHeader := r.Header.Get(timeoutHeaderName)
	t, err := strconv.Atoi(timeoutHeader)
	if err != nil || t <= 0 {
		t = 30
	}
	timeout := time.Duration(t) * time.Second
	session.SetTimeout(timeout)

	// Parse proxy
	proxy := r.Header.Get(proxyHeaderName)
	session.SetProxy(proxy)

	var body any
	if r.Method == fhttp.MethodPost {
		body = r.Body
	} else {
		body = nil
	}

	req := &azuretls.Request{
		Method:           r.Method,
		Url:              urlHeader,
		DisableRedirects: !allowRedirects,
		IgnoreBody:       true,
		Body:             body,
	}

	return session, req, nil
}

// SetHeaders sets the custom headers received in the server to the session
func SetHeaders(s *azuretls.Session, headers fhttp.Header) {
	browserHeaders := browser.Chrome126
	customHeaderNames := []string{
		urlHeaderName,
		proxyHeaderName,
		redirectHeaderName,
		timeoutHeaderName,
		bufferingHeaderName,
	}
Outer:
	for k, v := range headers {
		for _, header := range customHeaderNames {
			if strings.ToLower(header) == strings.ToLower(k) {
				continue Outer
			}
		}

		exist := browserHeaders.Get(strings.ToLower(k)) != ""
		if !exist {
			browserHeaders = append(browserHeaders, []string{k, v[0]})
		}
	}

	s.OrderedHeaders = browserHeaders
}

func SetCookies(url_ string, s *azuretls.Session, c []*fhttp.Cookie) {
    parsed, err := url.Parse(url_)
    if err != nil {
        return
    }

    s.CookieJar.SetCookies(parsed, c)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value

	}
	return fallback
}
