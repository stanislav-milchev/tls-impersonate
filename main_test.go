package main

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	http "github.com/Noooste/fhttp"
	"github.com/stretchr/testify/assert"
)

type apiResponse struct {
	Http string `json:"http_version"`
	Tls  struct {
		Ja3 string `json:"ja3"`
	}
}

type mockResponseWriter struct {
	statusCode int
	headers    http.Header
	body       *bytes.Buffer
}

func NewMockResponseWriter(header http.Header, body *bytes.Buffer, statusCode int) *mockResponseWriter {
	return &mockResponseWriter{
		statusCode: statusCode,
		headers:    header,
		body:       body,
	}
}

func (m *mockResponseWriter) Header() http.Header {
	return m.headers
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

func (m *mockResponseWriter) Write(data []byte) (int, error) {
	return m.body.Write(data)
}

func TestHandleReq(t *testing.T) {
	url := "https://tls.peet.ws/api/all"
	headers := make(http.Header)
	headers["x-tls-url"] = []string{"https://tls.peet.ws/api/all"}

	r, err := http.NewRequest(
		http.MethodGet,
		url,
		http.NoBody,
	)

	if err != nil {
		t.Fatal(err)
	}

	for k, v := range headers {
		r.Header.Set(k, v[0])
	}

	w := NewMockResponseWriter(make(http.Header), &bytes.Buffer{}, 0)

	HandleReq(w, r)
	var apiR apiResponse
	err = json.Unmarshal(w.body.Bytes(), &apiR)

	if err != nil {
		t.Fatal(err)
	}

	expectedJa3 := "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,10-35-0-65281-27-45-43-16-17513-18-51-11-13-23-5-65037,25497-29-23-24,0"

	assert.Equal(t, sortJa3(expectedJa3), sortJa3(apiR.Tls.Ja3))
}

func sortJa3(ja3 string) string {
	groups := strings.Split(ja3, ",")
	var sortedJa3 []string

	for _, g := range groups {
		extensions := strings.Split(g, "-")
		slices.Sort(extensions)
		sorted := strings.Join(extensions, "-")

		sortedJa3 = append(sortedJa3, sorted)
	}

	return strings.Join(sortedJa3, ",")
}

// Request tls peet to check if fingerpirnt matches
func TestNewRequest(t *testing.T) {
	url := "https://tls.peet.ws/api/all"

	headers := make(http.Header)
	headers["x-tls-url"] = []string{"https://tls.peet.ws/api/all"}

	r, err := http.NewRequest(
		http.MethodGet,
		url,
		http.NoBody,
	)

	if err != nil {
		t.Fatal(err)
	}

	for k, v := range headers {
		r.Header.Set(k, v[0])
	}

	_, req, err := NewRequest(r)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, url, req.Url)
	assert.Equal(t, r.Method, req.Method)
	assert.Empty(t, req.Body)

	//response, err := s.Get("https://tls.peet.ws/api/all")

}

// Request builder
