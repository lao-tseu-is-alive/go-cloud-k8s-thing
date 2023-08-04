package main

import (
	"encoding/json"
	"fmt"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/gohttpclient"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/url"
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
	httpMethod     string
	url            string
	useJwtToken    bool
	body           string
}

// TestMainExec is instantiating the "real" main code using the env variable (in your .env files if you use the Makefile rule)
func TestMainExec(t *testing.T) {
	listenPort, err := config.GetPortFromEnv(defaultPort)
	if err != nil {
		t.Errorf("💥💥 ERROR: 'calling GetPortFromEnv got error: %v'\n", err)
		return
	}
	listenAddr := fmt.Sprintf("http://localhost%s", listenPort)
	fmt.Printf("INFO: 'Will start HTTP server listening on port %s'\n", listenAddr)

	newRequest := func(method, url string, body string) *http.Request {
		fmt.Printf("INFO: 🚀🚀'newRequest %s on %s ##BODY : %+v'\n", method, url, body)
		r, err := http.NewRequest(method, url, strings.NewReader(body))
		if err != nil {
			t.Fatalf("### ERROR http.NewRequest %s on [%s] error is :%v\n", method, url, err)
		}
		if method == http.MethodPost {
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		return r
	}

	// preparing for testing a pseudo-valid authentication
	formLogin := make(url.Values)
	formLogin.Set("login", defaultUsername)
	formLogin.Set("pass", defaultFakeStupidPass)

	// preparing for testing an invalid authentication
	formLoginWrong := make(url.Values)
	formLoginWrong.Set("login", defaultUsername)
	formLoginWrong.Set("pass", "anObviouslyWrongPass")

	tests := []testStruct{
		{
			name:           "GET /  should contain html tag",
			wantStatusCode: http.StatusOK,
			contentType:    MIMEHtmlCharsetUTF8,
			wantBody:       "<html",
			paramKeyValues: make(map[string]string, 0),
			httpMethod:     http.MethodGet,
			url:            "/",
			useJwtToken:    false,
			body:           "",
		},
		{
			name:           "POST / should return an http error method not allowed ",
			wantStatusCode: http.StatusMethodNotAllowed,
			contentType:    MIMEHtmlCharsetUTF8,
			wantBody:       "Method Not Allowed",
			paramKeyValues: make(map[string]string, 0),
			httpMethod:     http.MethodPost,
			url:            "/",
			useJwtToken:    false,
			body:           `{"junk":"test with junk text"}`,
		},
		{
			name:           "GET /aroutethatwillneverexisthere should return an http error not found ",
			wantStatusCode: http.StatusNotFound,
			contentType:    MIMEHtmlCharsetUTF8,
			wantBody:       "page not found",
			paramKeyValues: make(map[string]string, 0),
			httpMethod:     http.MethodGet,
			url:            "/aroutethatwillneverexisthere",
			useJwtToken:    false,
			body:           "",
		},
		{
			name:           "POST /login with valid credential should return a JWT token ",
			wantStatusCode: http.StatusOK,
			contentType:    MIMEHtmlCharsetUTF8,
			wantBody:       "TOKEN",
			paramKeyValues: make(map[string]string, 0),
			httpMethod:     http.MethodPost,
			url:            "/login",
			useJwtToken:    false,
			body:           formLogin.Encode(),
		},
		{
			name:           "POST /login with invalid credential should return an error ",
			wantStatusCode: http.StatusUnauthorized,
			contentType:    MIMEHtmlCharsetUTF8,
			wantBody:       "unauthorized request: username not found or invalid password",
			paramKeyValues: make(map[string]string, 0),
			httpMethod:     http.MethodPost,
			url:            "/login",
			useJwtToken:    false,
			body:           formLoginWrong.Encode(),
		},
		{
			name:           "GET /thing without JWT token should return an error",
			wantStatusCode: http.StatusUnauthorized,
			contentType:    MIMEHtmlCharsetUTF8,
			wantBody:       "missing or malformed jwt",
			paramKeyValues: make(map[string]string, 0),
			httpMethod:     http.MethodGet,
			url:            defaultSecuredApi + "/thing",
			useJwtToken:    false,
			body:           formLoginWrong.Encode(),
		},
		{
			name:           "GET /thing with valid JWT token should return an list of Things",
			wantStatusCode: http.StatusOK,
			contentType:    MIMEHtmlCharsetUTF8,
			wantBody:       "created_at",
			paramKeyValues: make(map[string]string, 0),
			httpMethod:     http.MethodGet,
			url:            defaultSecuredApi + "/thing?limit=1&type=2&created_by=999",
			useJwtToken:    true,
			body:           formLoginWrong.Encode(),
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

	// let's get first a valid JWT TOKEN
	req := newRequest(http.MethodPost, listenAddr+"/login", formLogin.Encode())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("###Problem requesting JWT TOKEN 💥💥 ERROR : %s\n%+v", err, resp)
		t.Fatal(err)
	}
	defer resp.Body.Close()
	receivedJson, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("###Problem reading JWT TOKEN 💥💥 ERROR : %s\n%+v", err, resp)
		t.Fatal(err)
	}
	fmt.Printf("TOKEN retrieved 💡👉 status : %v, response.Body:\n%s\n", resp.StatusCode, string(receivedJson))
	type JwtToken struct {
		TOKEN string
	}
	var myToken JwtToken
	err = json.Unmarshal(receivedJson, &myToken)
	if err != nil {
		fmt.Printf("###Problem Unmarshalling JWT TOKEN 💥💥 ERROR : %s\n", err)
		t.Fatal(err)
	}
	fmt.Printf("TOKEN=%s\n", myToken.TOKEN)
	// Create a Bearer string by appending string access token
	var bearer = "Bearer " + myToken.TOKEN

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := newRequest(tt.httpMethod, listenAddr+tt.url, tt.body)
			if tt.useJwtToken {
				r.Header.Add("Authorization", bearer)
			}
			resp, err := http.DefaultClient.Do(r)
			if DEBUG {
				fmt.Printf("### %s : %s on %s\n", tt.name, r.Method, r.URL)
			}
			if err != nil {
				fmt.Printf("### GOT ERROR : %s\n%+v", err, resp)
				t.Fatal(err)
			}
			defer resp.Body.Close()
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
