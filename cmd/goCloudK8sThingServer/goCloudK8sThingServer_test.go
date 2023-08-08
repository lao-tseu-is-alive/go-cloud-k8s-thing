package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/gohttpclient"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/tools"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-thing/pkg/version"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	DEBUG                           = false
	assertCorrectStatusCodeExpected = "expected status code should be returned"
	newThingId                      = "24466b0c-686d-42a3-87ef-bf6cefeb3d35"
	newThingExternalId              = "1234567890"
	exampleThing                    = `
{
    "created_by": 999999,
    "description": "La belle ville de 'Château Français'' de l'école œcuménique des chevaux & exemple de caractère",
    "comment": "🌠✮  🎀  𝒰𝓃 𝑒𝓍𝑒𝓂𝓅𝓁𝑒 𝒹𝑒 𝓉𝑒𝓍𝓉𝑒 𝒶𝓋𝑒𝒸 𝒹𝑒𝓈 𝒸𝒶𝓇𝒶𝒸𝓉è𝓇𝑒𝓈 𝒰𝓃𝒾𝒸❀𝒹𝑒  🎀  ✮🌠 🎁📣❤️ 💔☀️🔥💰⏰💥✊📢🎯👥🆕👩‍🔧👨‍💼👩‍💼🕶👓🎩🎓☄️⛳️ 𝑻𝒉𝒆 𝒒𝒖𝒊𝒄𝒌 𝒃𝒓𝒐𝒘𝒏 𝒇𝒐𝒙 𝒋𝒖𝒎𝒑𝒔 𝒐𝒗𝒆𝒓 𝒕𝒉𝒆 𝒍𝒂𝒛𝒚 𝒅𝒐𝒈",
    "external_id": 1234567890,
    "id": "24466b0c-686d-42a3-87ef-bf6cefeb3d35",
    "inactivated": false,
    "name": "Château Français",
    "pos_x": 2537607.64,
    "pos_y": 1152609.12,
    "type_id": 2,
    "validated": false
  }
`
	exampleThingUpdate = `
{
    "description": "La belle ville de 'Château Français'' de l'école œcuménique des chevaux & exemple de caractère",
    "external_id": 1234567890,
    "id": "24466b0c-686d-42a3-87ef-bf6cefeb3d35",
    "inactivated": false,
    "name": "Château Français",
    "pos_x": 2537607.64,
    "pos_y": 1152609.12,
    "type_id": 2,
    "validated": false,
	"comment": "À Noël la livraison de maïs, surtout après un Æquinoxe vernal est aussi hypothétique que la floraison des æschynanthes qui n'apparaîtra que dans l'Œil d'un cyclone métaphysique "
  }
`
)

type testStruct struct {
	name                         string
	contentType                  string
	wantStatusCode               int
	wantBody                     string
	paramKeyValues               map[string]string
	httpMethod                   string
	url                          string
	useFormUrlencodedContentType bool
	useJwtToken                  bool
	body                         string
}

func MakeHttpRequest(method, url, sendBody, token string, caCert []byte, l golog.MyLogger, defaultReadTimeout time.Duration) (string, error) {
	// Create a Bearer string by appending string access token
	var bearer = "Bearer " + token

	// Create a new request using http
	req, err := http.NewRequest("GET", url, nil)

	// add authorization header to the req
	req.Header.Add("Authorization", bearer)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}
	// Send req using http Client
	client := &http.Client{
		Transport: tr,
		Timeout:   defaultReadTimeout,
	}
	resp, err := client.Do(req)

	if err != nil {
		l.Error("GetJsonFromUrlWithBearerAuth: Error on response.\n[ERROR] -", err)
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			l.Error("GetJsonFromUrlWithBearerAuth: Error on Body.Close().\n[ERROR] -", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error("GetJsonFromUrlWithBearerAuth: Error while reading the response bytes:", err)
		return "", err
	}
	return string(body), nil
}

// TestMainExec is instantiating the "real" main code using the env variable (in your .env files if you use the Makefile rule)
func TestMainExec(t *testing.T) {
	prefix := fmt.Sprintf("%s_TESTING ", version.APP)
	l, err := golog.NewLogger("zap", golog.DebugLevel, prefix)
	listenPort, err := config.GetPortFromEnv(defaultPort)
	if err != nil {
		t.Errorf("💥💥 ERROR: 'calling GetPortFromEnv got error: %v'\n", err)
		return
	}
	listenAddr := fmt.Sprintf("http://localhost%s", listenPort)
	fmt.Printf("INFO: 'Will start HTTP server listening on port %s'\n", listenAddr)

	newRequest := func(method, url string, body string, useFormUrlencodedContentType bool) *http.Request {
		fmt.Printf("INFO: 🚀🚀'newRequest %s on %s ##BODY : %+v'\n", method, url, body)
		r, err := http.NewRequest(method, url, strings.NewReader(body))
		if err != nil {
			t.Fatalf("### ERROR http.NewRequest %s on [%s] error is :%v\n", method, url, err)
		}
		if method == http.MethodPost && useFormUrlencodedContentType {
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		} else {
			r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		}
		return r
	}
	// preparing for testing a pseudo-valid authentication
	formLogin := make(url.Values)
	formLogin.Set("login", defaultUsername)
	formLogin.Set("pass", defaultFakeStupidPass)

	getValidToken := func() string {
		// let's get first a valid JWT TOKEN
		req := newRequest(http.MethodPost, listenAddr+"/login", formLogin.Encode(), true)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("###getValidToken: Problem requesting JWT TOKEN 💥💥 ERROR : %s\n%+v", err, resp)
			t.Fatal(err)
		}
		defer resp.Body.Close()
		receivedJson, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("###getValidToken: Problem reading JWT TOKEN 💥💥 ERROR : %s\n%+v", err, resp)
			t.Fatal(err)
		}
		fmt.Printf("getValidToken: TOKEN retrieved 💡👉 status : %v, response.Body:\n%s\n", resp.StatusCode, string(receivedJson))
		type JwtToken struct {
			TOKEN string
		}
		var myToken JwtToken
		err = json.Unmarshal(receivedJson, &myToken)
		if err != nil {
			fmt.Printf("###getValidToken: Problem Unmarshalling JWT TOKEN 💥💥 ERROR : %s\n", err)
			t.Fatal(err)
		}
		fmt.Printf("TOKEN=\"%s\"\n", myToken.TOKEN)
		return myToken.TOKEN
	}

	// preparing for testing an invalid authentication
	formLoginWrong := make(url.Values)
	formLoginWrong.Set("login", defaultUsername)
	formLoginWrong.Set("pass", "anObviouslyWrongPass")

	dbDsn, err := config.GetPgDbDsnUrlFromEnv(defaultDBIp, defaultDBPort,
		tools.ToSnakeCase(version.APP), version.AppSnake, defaultDBSslMode)
	if err != nil {
		t.Fatalf("💥💥 error doing config.GetPgDbDsnUrlFromEnv. error: %v\n", err)
	}
	db, err := database.GetInstance("pgx", dbDsn, runtime.NumCPU(), l)
	if err != nil {
		t.Fatalf("💥💥 error doing users.GetPgxConn(postgres, dbDsn  : %v\n", err)
	}
	defer db.Close()

	// checking database connection
	dbVersion, err := db.GetVersion()
	if err != nil {
		t.Fatalf("💥💥 error doing dbConn.GetVersion() error: %v", err)
	}
	fmt.Printf("connected to db version : %s", dbVersion)
	// removing latest test record if exist
	count, err := db.GetQueryInt("SELECT COUNT(*) FROM go_thing.thing WHERE id = $1;", newThingId)
	if err != nil {
		t.Fatalf("problem during cleanup before test DB. failed db.Query err: %v", err)
	}
	if count > 0 {
		fmt.Printf(" This Id(%v) does exist  will cleanup before running test", newThingId)
		db.ExecActionQuery("DELETE FROM  go_thing.thing WHERE id=$1", newThingId)
	}

	tests := []testStruct{
		{
			name:                         "GET /  should contain html tag",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "<html",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodGet,
			url:                          "/",
			useFormUrlencodedContentType: false,
			useJwtToken:                  false,
			body:                         "",
		},
		{
			name:                         "POST / should return an http error method not allowed ",
			wantStatusCode:               http.StatusMethodNotAllowed,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "Method Not Allowed",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPost,
			url:                          "/",
			useFormUrlencodedContentType: true,
			useJwtToken:                  false,
			body:                         `{"junk":"test with junk text"}`,
		},
		{
			name:                         "GET /aroutethatwillneverexisthere should return an http error not found ",
			wantStatusCode:               http.StatusNotFound,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "page not found",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodGet,
			url:                          "/aroutethatwillneverexisthere",
			useFormUrlencodedContentType: false,
			useJwtToken:                  false,
			body:                         "",
		},
		{
			name:                         "POST /login with valid credential should return a JWT token ",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "TOKEN",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPost,
			url:                          "/login",
			useFormUrlencodedContentType: true,
			useJwtToken:                  false,
			body:                         formLogin.Encode(),
		},
		{
			name:                         "POST /login with invalid credential should return an error ",
			wantStatusCode:               http.StatusUnauthorized,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "unauthorized request: username not found or invalid password",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPost,
			url:                          "/login",
			useFormUrlencodedContentType: true,
			useJwtToken:                  false,
			body:                         formLoginWrong.Encode(),
		},
		{
			name:                         "GET /thing without JWT token should return an error",
			wantStatusCode:               http.StatusUnauthorized,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "missing or malformed jwt",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/thing",
			useFormUrlencodedContentType: false,
			useJwtToken:                  false,
			body:                         formLoginWrong.Encode(),
		},
		{
			name:                         "POST /thing with valid JWT token should create a new Things",
			wantStatusCode:               http.StatusCreated,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "create_at",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPost,
			url:                          defaultSecuredApi + "/thing",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         exampleThing,
		},
		{
			name:                         "POST /thing with id already present should return error",
			wantStatusCode:               http.StatusBadRequest,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "already exist",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPost,
			url:                          defaultSecuredApi + "/thing",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         exampleThing,
		},
		{
			name:                         "GET /thing with valid JWT token should return an list of Things",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "created_at",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/thing?limit=1&offset=0&type=2&created_by=999",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         formLoginWrong.Encode(),
		},
		{
			name:                         "GET /thing with existing id should return the valid thing",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "\"id\":\"" + newThingId + "\"",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/thing/" + newThingId,
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "GET /thing with non-existing id should return StatusNotFound",
			wantStatusCode:               http.StatusNotFound,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/thing/11111111-4444-5555-6666-777777777777",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "PUT /thing with existing id should return the valid thing",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "\"comment\":\"À Noël ",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPut,
			url:                          defaultSecuredApi + "/thing/" + newThingId,
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         exampleThingUpdate,
		},
		{
			name:                         "PUT /thing with non-existing id should return StatusNotFound",
			wantStatusCode:               http.StatusNotFound,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "cannot update this id, it does not exist",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPut,
			url:                          defaultSecuredApi + "/thing/11111111-4444-5555-6666-777777777777",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "GET /thing/by-external-id with existing id should return the valid thing",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "\"id\":\"" + newThingId + "\"",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/thing/by-external-id/" + newThingExternalId + "?limit=1&offset=0",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "GET /thing/by-external-id with non-existing id should return StatusNotFound",
			wantStatusCode:               http.StatusNotFound,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "[]", // should return an empty array
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/thing/by-external-id/2147483645",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "DELETE /thing with existing id should return StatusNoContent",
			wantStatusCode:               http.StatusNoContent,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodDelete,
			url:                          defaultSecuredApi + "/thing/" + newThingId,
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "DELETE /thing with non-existing id should return StatusNotFound",
			wantStatusCode:               http.StatusNotFound,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodDelete,
			url:                          defaultSecuredApi + "/thing/11111111-4444-5555-6666-777777777777",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
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

	// Create a Bearer string by appending JWT string access token
	var bearer = "Bearer " + getValidToken()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare the request for this test case
			r := newRequest(tt.httpMethod, listenAddr+tt.url, tt.body, tt.useFormUrlencodedContentType)
			// add the JWT token if asked
			if tt.useJwtToken {
				r.Header.Add("Authorization", bearer)
			}
			if DEBUG {
				fmt.Printf("### %s : will try %s on %s\n", tt.name, r.Method, r.URL)
			}
			resp, err := http.DefaultClient.Do(r)
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
