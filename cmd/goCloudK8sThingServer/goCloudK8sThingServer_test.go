package main

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/gohttpclient"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/tools"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-thing/pkg/thing"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-thing/pkg/version"
	"github.com/stretchr/testify/assert"
)

const (
	DEBUG                           = true
	assertCorrectStatusCodeExpected = "expected status code should be returned"
	urlLogin                        = "/login"
	urlThing                        = "/thing"
	urlTypeThing                    = "/types"
	newThingId                      = "24466b0c-686d-42a3-87ef-bf6cefeb3d35"
	urlNewThingId                   = "/thing/" + newThingId
	bodyIdNewThing                  = "\"id\":\"" + newThingId + "\""
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
    "validated": true,
    "created_by": 999999,
	"comment": "À Noël la livraison de maïs, surtout après un Æquinoxe vernal est aussi hypothétique que la floraison des æschynanthes qui n'apparaîtra que dans l'Œil d'un cyclone métaphysique "
  }
`
	exampleTypeThing = `
{
  "id": 99999,
  "description": "Piscine publique ou privée",
  "geometry_type": "bbox", 
  "external_id": 987654321,
  "name": "Piscine",
  "icon_path": "/img/gomarker_star_blue.png",
  "created_by": 1
}

`
	exampleTypeThingUpdate = `
{
  "description": "Piscine publique ou privée, avec parfois des Dischidia nummularia ",
  "comment": "Attention la piscine est 🤔 ...️⁉️⚠️  🏴 ‍☠️ 💀 ☠️ ☢️ ☣️ 💣 💥, ",
  "geometry_type": "bbox", 
  "external_id": 987654321,
  "name": "Piscine",
  "icon_path": "/img/gomarker_star_blue.png"
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

func MakeHttpRequest(method, url, sendBody, token string, caCert []byte, l *slog.Logger, defaultReadTimeout time.Duration) (string, error) {
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
		l.Error("GetJsonFromUrlWithBearerAuth: error on response.", "error", err)
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			l.Error("GetJsonFromUrlWithBearerAuth: error on body.Close()", "error", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error("GetJsonFromUrlWithBearerAuth: error while reading the response bytes:", "error", err)
		return "", err
	}
	return string(body), nil
}

// TestMainExec is instantiating the "real" main code using the env variable (in your .env files if you use the Makefile rule)
func TestMainExec(t *testing.T) {
	l := golog.NewLogger("simple", os.Stdout, golog.DebugLevel, version.APP)
	listenPort, _ := config.GetPort(defaultPort)
	listenAddr := fmt.Sprintf("http://localhost:%d", listenPort)
	fmt.Printf("INFO: 'Will start HTTP server listening on port %s'\n", listenAddr)
	// common messages
	//nameCannotBeEmpty := fmt.Sprintf("name: value length must be at least 5field %s cannot be empty", "name")
	//nameCannotBeEmpty := fmt.Sprintf("thing.%s: value length must be at least 5 characters ", "name")
	nameMinLengthMsg := fmt.Sprintf("%s: value length must be at least %d", "name", thing.MinNameLength)
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
	// set local admin user for test
	adminUsername, err := config.GetAdminUser("goadmin")
	if err != nil {
		l.Error("💥💥 error retrieving admin user", "error", err)
		os.Exit(1)
	}
	adminPassword, err := config.GetAdminPassword()
	if err != nil {
		l.Error("💥💥 error retrieving admin password", "error", err)
		os.Exit(1)
	}
	l.Warn("will try to connect with admin credentials", "user", adminUsername, "password", adminPassword)
	h := sha256.New()
	h.Write([]byte(adminPassword))
	adminPasswordHash := fmt.Sprintf("%x", h.Sum(nil))
	// preparing for testing a pseudo-valid authentication
	formLogin := make(url.Values)
	formLogin.Set("login", adminUsername)
	formLogin.Set("hashed", adminPasswordHash)

	getValidToken := func() string {
		// let's get first a valid JWT TOKEN

		req := newRequest(http.MethodPost, listenAddr+urlLogin, formLogin.Encode(), true)
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
	formLoginWrong.Set("login", adminUsername)
	formLoginWrong.Set("pass", "anObviouslyWrongPass")

	dbDsn, err := config.GetPgDbDsnUrl(defaultDBIp, defaultDBPort,
		tools.ToSnakeCase(version.APP), version.AppSnake, defaultDBSslMode)
	if err != nil {
		l.Error("💥💥 error getting database DSN", "error", err)
		os.Exit(1)
	}
	dbConnCtx, dbConnCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbConnCancel()
	db, err := database.GetInstance(dbConnCtx, "pgx", dbDsn, runtime.NumCPU(), l)
	if err != nil {
		l.Error("💥💥 error doing database.GetInstance", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// checking database connection
	dbVersion, err := db.GetVersion(context.Background())
	if err != nil {
		l.Error("💥💥 error doing dbConn.GetVersion", "error", err)
		os.Exit(1)
	}
	l.Info("connected to db", "version", dbVersion)
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()
	existTable, err := db.GetQueryBool(dbCtx, "SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'go_thing' AND tablename  = 'thing');")
	if err != nil {
		t.Fatalf("problem verifying if thing exist in DB. failed db.Query err: %v", err)
	}
	if existTable {
		// removing latest test record if exist only if the thing table already exist
		count, err := db.GetQueryInt(dbCtx, "SELECT COUNT(*) FROM go_thing.thing WHERE id = $1;", newThingId)
		if err != nil {
			t.Fatalf("problem during cleanup before test DB. failed db.Query err: %v", err)
		}
		if count > 0 {
			fmt.Printf(" This Id(%v) does exist  will cleanup before running test", newThingId)
			db.ExecActionQuery(dbCtx, "DELETE FROM  go_thing.thing WHERE id=$1", newThingId)
		}
	}

	// deleting type thing of previous run if it's still present and if table type_thing exist only
	existTableTypeThing, err := db.GetQueryBool(dbCtx, "SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'go_thing' AND tablename  = 'type_thing');")
	if err != nil {
		t.Fatalf("problem verifying if thing exist in DB. failed db.Query err: %v", err)
	}
	var existingMaxTypeThingId = 112
	var existingCountTypeThingId = 107
	if existTableTypeThing {
		sqlDeleteInsertedTypeThing := "DELETE FROM go_thing.type_thing WHERE external_id=987654321;"
		_, err = db.ExecActionQuery(dbCtx, sqlDeleteInsertedTypeThing)
		if err != nil {
			t.Fatalf("problem trying to delete type_thing from previous test doing cleanup before running tests. failed db.Query err: %v", err)
		}
		typeThingMaxIdSql := "SELECT MAX(id) FROM go_thing.type_thing"
		existingMaxTypeThingId, err = db.GetQueryInt(dbCtx, typeThingMaxIdSql)
		if err != nil {
			t.Fatalf("problem trying to retrieve max id for typeThing cleanup before running test. failed db.Query err: %v", err)
		}
		resetSequence := "SELECT setval('go_thing.type_thing_id_seq', max(id)) FROM go_thing.type_thing;"
		_, err = db.ExecActionQuery(dbCtx, resetSequence)
		if err != nil {
			t.Fatalf("problem trying to resetSequence to max id for type_thing_id_seq while doing cleanup before running tests. failed db.Query err: %v", err)
		}
	}
	// incrementing one to get the real id of insert
	existingMaxTypeThingId += 1
	urTypeThingNewId := "/types/" + strconv.Itoa(existingMaxTypeThingId)

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
			wantBody:                     "token",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPost,
			url:                          urlLogin,
			useFormUrlencodedContentType: true,
			useJwtToken:                  false,
			body:                         formLogin.Encode(),
		},
		{
			name:                         "POST /login with invalid credential should return an error ",
			wantStatusCode:               http.StatusUnauthorized,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "username not found or password invalid",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPost,
			url:                          urlLogin,
			useFormUrlencodedContentType: true,
			useJwtToken:                  false,
			body:                         formLoginWrong.Encode(),
		},
		{
			name:                         "GET /thing without JWT token should return an error",
			wantStatusCode:               http.StatusUnauthorized,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "missing authorization header",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + urlThing,
			useFormUrlencodedContentType: false,
			useJwtToken:                  false,
			body:                         "",
		},
		{
			name:                         "POST /types with empty name should return bad request",
			wantStatusCode:               http.StatusBadRequest,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     nameMinLengthMsg,
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPost,
			url:                          defaultSecuredApi + urlTypeThing,
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         `{ "geometry_type": "bbox",    "external_id": 987654321,   "name": " " } `,
		},
		{
			name:                         "POST /types with name shorter then 5 chars should return bad request",
			wantStatusCode:               http.StatusBadRequest,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     nameMinLengthMsg,
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPost,
			url:                          defaultSecuredApi + urlTypeThing,
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         `{ "geometry_type": "bbox",    "external_id": 987654321,   "name": "tutu" } `,
		},
		{
			name:                         "POST /types with valid JWT token should create a new typeThings",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "createdBy",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPost,
			url:                          defaultSecuredApi + urlTypeThing,
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         exampleTypeThing,
		},
		{
			name:                         "POST /thing with empty name should return bad request",
			wantStatusCode:               http.StatusBadRequest,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     nameMinLengthMsg,
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPost,
			url:                          defaultSecuredApi + urlThing,
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         `{ "created_by": 999999, "external_id": 1234567890,     "id": "34466b0c-686d-42a3-87ef-bf6cefeb3d35","name": " ",     "pos_x": 2537607.64,     "pos_y": 1152609.12,     "type_id": 2,     "validated": false   } `,
		},
		{
			name:                         "POST /thing with name shorter then 5 chars should return bad request",
			wantStatusCode:               http.StatusBadRequest,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     nameMinLengthMsg,
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPost,
			url:                          defaultSecuredApi + urlThing,
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         ` { "created_by": 999999, "external_id": 1234567890,     "id": "34466b0c-686d-42a3-87ef-bf6cefeb3d35","name": "toto",     "pos_x": 2537607.64,     "pos_y": 1152609.12,     "type_id": 2,     "validated": false   } `,
		},
		{
			name:                         "POST /thing with valid JWT token should create a new Things",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "createdAt",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPost,
			url:                          defaultSecuredApi + urlThing,
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         exampleThing,
		},
		{
			name:                         "POST /thing with id already present should return error",
			wantStatusCode:               http.StatusConflict,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "already exist",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPost,
			url:                          defaultSecuredApi + urlThing,
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         exampleThing,
		},
		{
			name:                         "GET /status with valid JWT token should return JWT user data",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "SimpleAdminAuthenticator_goadmin",
			paramKeyValues:               make(map[string]string),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/status",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "GET /thing with valid JWT token should return an list of Things",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     bodyIdNewThing,
			paramKeyValues:               make(map[string]string),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/thing?limit=1&offset=0&type=2&created_by=999999",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "GET /thing with created_by and validated false should return the valid thing",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     bodyIdNewThing,
			paramKeyValues:               make(map[string]string),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/thing?limit=1&offset=0&type=2&created_by=999999&validated=false",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "GET /thing with created_by and validated true should return empty list",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "{}",
			paramKeyValues:               make(map[string]string),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/thing?limit=1&offset=0&type=2&created_by=999999&validated=true",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "GET /thing with existing id should return the valid thing",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     bodyIdNewThing,
			paramKeyValues:               make(map[string]string),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + urlNewThingId,
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "GET /thing with non-existing id should return StatusNotFound",
			wantStatusCode:               http.StatusNotFound,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "",
			paramKeyValues:               make(map[string]string),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/thing/11111111-4444-5555-6666-777777777777",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "GET /thing/count without filter params should return an int > 1",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "1",
			paramKeyValues:               make(map[string]string),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/thing/count",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "PUT /thing with existing id but empty name should return bad request",
			wantStatusCode:               http.StatusBadRequest,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     nameMinLengthMsg,
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPut,
			url:                          defaultSecuredApi + urlNewThingId,
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         `{ "created_by": 999999, "external_id": 1234567890,     "id": "34466b0c-686d-42a3-87ef-bf6cefeb3d35","name": " ",     "pos_x": 2537607.64,     "pos_y": 1152609.12,     "type_id": 2,     "validated": false   } `,
		},
		{
			name:                         "PUT /thing with existing id but name shorter then 5 chars should return bad request",
			wantStatusCode:               http.StatusBadRequest,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     nameMinLengthMsg,
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPut,
			url:                          defaultSecuredApi + urlNewThingId,
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         `{ "created_by": 999999, "external_id": 1234567890,     "id": "34466b0c-686d-42a3-87ef-bf6cefeb3d35","name": "titi",     "pos_x": 2537607.64,     "pos_y": 1152609.12,     "type_id": 2,     "validated": false   } `,
		},

		{
			name:                         "PUT /thing with existing id should return the valid thing",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "\"comment\":\"À Noël ",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPut,
			url:                          defaultSecuredApi + urlNewThingId,
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         exampleThingUpdate,
		},
		{
			name:                         "PUT /thing with non-existing id should return StatusNotFound",
			wantStatusCode:               http.StatusNotFound,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "not found",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPut,
			url:                          defaultSecuredApi + "/thing/11111111-4444-5555-6666-777777777777",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         exampleThingUpdate,
		},
		{
			name:                         "GET /thing/by-external-id with existing id should return the valid thing",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     bodyIdNewThing,
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
			wantBody:                     "",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/thing/by-external-id/2147483645",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "GET /thing/search with existing keyword should return the valid thing",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     bodyIdNewThing,
			paramKeyValues:               make(map[string]string),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/thing/search?limit=1&offset=0&keywords=æschynanthes",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "GET /thing/search with existing keyword and validated should return the valid thing",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     bodyIdNewThing,
			paramKeyValues:               make(map[string]string),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/thing/search?limit=1&offset=0&keywords=æschynanthes&validated=true",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "GET /thing/search with existing keyword and validated false should return thing",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "things",
			paramKeyValues:               make(map[string]string),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/thing/search?limit=1&offset=0&keywords=æschynanthes&validated=false",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "GET /thing/search validated should return the valid thing",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     bodyIdNewThing,
			paramKeyValues:               make(map[string]string),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/thing/search?limit=1&offset=0&validated=true",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "GET /thing/search with created_by should return the valid thing",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     bodyIdNewThing,
			paramKeyValues:               make(map[string]string),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/thing/search?limit=1&offset=0&created_by=999999",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "GET /thing/search with non-existing keyword should return ok and empty object",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "{}",
			paramKeyValues:               make(map[string]string),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/thing/search?limit=1&offset=0&keywords=anticonstitutionnellement",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "DELETE /thing with existing id should return StatusOK",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodDelete,
			url:                          defaultSecuredApi + urlNewThingId,
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
		// TypeThing scenarios
		{
			name:                         "GET /types without JWT token should return an error",
			wantStatusCode:               http.StatusUnauthorized,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "missing authorization header",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + urlTypeThing,
			useFormUrlencodedContentType: false,
			useJwtToken:                  false,
			body:                         "",
		},
		{
			name:                         "GET /types with valid JWT token should return a list of TypeThings",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "createdAt",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/types?limit=1&offset=0&created_by=999999",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "GET /types with existing id should return the valid typeThing",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "\"id\":" + strconv.Itoa(existingMaxTypeThingId),
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + urTypeThingNewId,
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "GET /types with non-existing id should return StatusNotFound",
			wantStatusCode:               http.StatusNotFound,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/types/8876541",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "PUT /types with empty name should return bad request",
			wantStatusCode:               http.StatusBadRequest,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     nameMinLengthMsg,
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPut,
			url:                          defaultSecuredApi + urTypeThingNewId,
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         ` { "geometry_type": "bbox",    "external_id": 987654321,   "name": " " } `,
		},
		{
			name:                         "PUT /types with name shorter then 5 chars should return bad request",
			wantStatusCode:               http.StatusBadRequest,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     nameMinLengthMsg,
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPut,
			url:                          defaultSecuredApi + urTypeThingNewId,
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         ` { "geometry_type": "bbox",    "external_id": 987654321,   "name": "zuzu" } `,
		},
		{
			name:                         "PUT /types with existing id should return the valid typeThing",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "\"comment\":\"Attention la piscine est 🤔",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPut,
			url:                          defaultSecuredApi + urTypeThingNewId,
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         strings.Replace(exampleTypeThingUpdate, "}", ",\n\"id\":"+strconv.Itoa(existingMaxTypeThingId)+"}", 1),
		},
		{
			name:                         "PUT /types with non-existing id should return StatusNotFound",
			wantStatusCode:               http.StatusNotFound,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "not found",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPut,
			url:                          defaultSecuredApi + "/types/8876541",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         strings.Replace(exampleTypeThingUpdate, "}", ",\n\"id\":8876541}", 1),
		},
		{
			name:                         "GET /types with with existing external_id should return the valid typeThing",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "\"id\":" + strconv.Itoa(existingMaxTypeThingId),
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/types?external_id=987654321&limit=1&offset=0",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "GET /types with with non-existing external_id should return empty object",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "{}",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/types?external_id=2147483645&limit=1&offset=0",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "GET /types with existing keyword should return the valid typeThing",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "\"id\":" + strconv.Itoa(6),
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/types?limit=1&offset=0&keywords=quai",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "GET /types with non-existing keyword should return empty object",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "{}",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/types?limit=1&offset=0&keywords=anticonstitutionnellement",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "GET /types/count should return the number of type thing id",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     strconv.Itoa(existingCountTypeThingId),
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodGet,
			url:                          defaultSecuredApi + "/types/count",
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "DELETE /types with existing id should return StatusOK",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodDelete,
			url:                          defaultSecuredApi + urTypeThingNewId,
			useFormUrlencodedContentType: false,
			useJwtToken:                  true,
			body:                         "",
		},
		{
			name:                         "DELETE /types with non-existing id should return StatusNotFound",
			wantStatusCode:               http.StatusNotFound,
			contentType:                  MIMEHtmlCharsetUTF8,
			wantBody:                     "",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodDelete,
			url:                          defaultSecuredApi + "/types/8876541",
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
	gohttpclient.WaitForHttpServer(listenAddr, 1*time.Second, 10, l)

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
