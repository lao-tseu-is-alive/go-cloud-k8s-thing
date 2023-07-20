package main

import (
	"fmt"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/gohttpclient"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	DEBUG                           = false
	assertCorrectStatusCodeExpected = "expected status code should be returned"
)

type testStruct struct {
	name           string
	contentType    string
	wantStatusCode int
	wantBody       string
	paramKeyValues map[string]string
	r              *http.Request
}

// TestMainExec is instantiating the "real" main code using the env variable (in your .env files if you use the Makefile rule)
func TestMainExec(t *testing.T) {
	listenAddr := fmt.Sprintf("http://localhost:%d/", defaultPort)
	err := os.Setenv("PORT", fmt.Sprintf("%d", defaultPort))
	if err != nil {
		t.Errorf("Unable to set env variable PORT")
		return
	}
	newRequest := func(method, url string, body string) *http.Request {
		r, err := http.NewRequest(method, listenAddr+url, strings.NewReader(body))
		if err != nil {
			t.Fatalf("### ERROR http.NewRequest %s on [%s] error is :%v\n", method, url, err)
		}
		return r
	}

	formLogin := make(url.Values)
	formLogin.Set("login", defaultUsername)
	formLogin.Set("pass", defaultFakeStupidPass)

	tests := []testStruct{
		{
			name:           "1: Get on default get handler should contain html tag",
			wantStatusCode: http.StatusOK,
			contentType:    MIMEHtmlCharsetUTF8,
			wantBody:       "<html",
			paramKeyValues: make(map[string]string, 0),
			r:              newRequest(http.MethodGet, "/", ""),
		},
		{
			name:           "2: Post on default get handler should return an http error method not allowed ",
			wantStatusCode: http.StatusMethodNotAllowed,
			contentType:    MIMEHtmlCharsetUTF8,
			wantBody:       "",
			paramKeyValues: make(map[string]string, 0),
			r:              newRequest(http.MethodPost, "/", `{"junk":"test with junk text"}`),
		},
		{
			name:           "3: Get on nonexistent route should return an http error not found ",
			wantStatusCode: http.StatusNotFound,
			contentType:    MIMEHtmlCharsetUTF8,
			wantBody:       "",
			paramKeyValues: make(map[string]string, 0),
			r:              newRequest(http.MethodGet, "/aroutethatwillneverexisthere", ""),
		},
		{
			name:           "4: POST to login with valid credential should return a JWT token ",
			wantStatusCode: http.StatusOK,
			contentType:    MIMEAppJSONCharsetUTF8,
			wantBody:       "token",
			paramKeyValues: make(map[string]string, 0),
			r:              newRequest(http.MethodPost, "/login", formLogin.Encode()),
		},
	}

	// starting main in his own go routine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		main()
	}()
	gohttpclient.WaitForHttpServer(listenAddr, 1*time.Second, 10)

	// let's test the default get handler
	resp, err := http.Get(listenAddr)
	if err != nil {
		t.Fatalf("Cannot make http get: %v\n", err)
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Should return an http status ok")

	receivedHtml, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %v\n", err)
	}
	assert.Contains(t, string(receivedHtml), "<html", "Response should contain the html tag.")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.Header.Set(HeaderContentType, tt.contentType)
			resp, err := http.DefaultClient.Do(tt.r)
			if DEBUG {
				fmt.Printf("### %s : %s on %s\n", tt.name, tt.r.Method, tt.r.URL)
			}
			defer resp.Body.Close()
			if err != nil {
				fmt.Printf("### GOT ERROR : %s\n%s", err, resp.Body)
				t.Fatal(err)
			}
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode, assertCorrectStatusCodeExpected)
			receivedJson, _ := io.ReadAll(resp.Body)

			if DEBUG {
				fmt.Printf("WANTED   :%T - %#v\n", tt.wantBody, tt.wantBody)
				fmt.Printf("RECEIVED :%T - %#v\n", receivedJson, string(receivedJson))
			}
			// check that receivedJson contains the specified tt.wantBody substring . https://pkg.go.dev/github.com/stretchr/testify/assert#Contains
			assert.Contains(t, string(receivedJson), tt.wantBody, "Response should contain what was expected.")
		})
	}
}
