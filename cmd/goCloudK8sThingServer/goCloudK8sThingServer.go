package main

import (
	"crypto/sha256"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cristalhq/jwt/v4"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goserver"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/metadata"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/tools"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-thing/pkg/thing"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-thing/pkg/version"
	"net/http"
	"runtime"
	"strings"
	"time"
)

const (
	defaultPort                = 9090
	defaultDBPort              = 5432
	defaultDBIp                = "127.0.0.1"
	defaultDBSslMode           = "prefer"
	defaultWebRootDir          = "goCloudK8sThingFront/dist/"
	defaultSqlDbMigrationsPath = "db/migrations"
	defaultSecuredApi          = "/goapi/v1"
	defaultThingAdminUsername  = "bill"
	charsetUTF8                = "charset=UTF-8"
	MIMEAppJSON                = "application/json"
	MIMEHtml                   = "text/html"
	MIMEAppJSONCharsetUTF8     = MIMEAppJSON + "; " + charsetUTF8
	MIMEHtmlCharsetUTF8        = MIMEHtml + "; " + charsetUTF8
	HeaderContentType          = "Content-Type"
)

// content holds our static web server content.
//
//go:embed goCloudK8sThingFront/dist/*
var content embed.FS

// sqlMigrations holds our db migrations sql files using https://github.com/golang-migrate/migrate
// in the line above you SHOULD have the same path  as const defaultSqlDbMigrationsPath
//
//go:embed db/migrations/*.sql
var sqlMigrations embed.FS

type ServiceThing struct {
	Log         golog.MyLogger
	dbConn      database.DB
	JwtSecret   []byte
	JwtDuration int
	adminUser   string
	adminHash   string
}

// UserLogin defines model for UserLogin.
type UserLogin struct {
	PasswordHash string `json:"password_hash"`
	Username     string `json:"username"`
}

// login is just a trivial stupid example to test this server
// you should use the jwt token returned from LoginUser  in github.com/lao-tseu-is-alive/go-cloud-k8s-user-group'
// and share the same secret with the above component
func (s ServiceThing) login(ctx echo.Context) error {
	s.Log.Debug("++ entering %v login()", ctx.Request().Method)
	uLogin := new(UserLogin)
	username := ctx.FormValue("login")
	fakePassword := ctx.FormValue("pass")
	s.Log.Debug("username: %s , password: %s", username, fakePassword)
	// maybe it was not a form but a fetch data post
	if len(strings.Trim(username, " ")) < 1 {
		if err := ctx.Bind(uLogin); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid user login or json format in request body")
		}
	} else {
		uLogin.Username = username
		uLogin.PasswordHash = fakePassword
	}
	s.Log.Debug("About to check username: %s , password: %s", uLogin.Username, uLogin.PasswordHash)
	// Throws unauthorized error
	if uLogin.Username != s.adminUser || uLogin.PasswordHash != s.adminHash {
		s.Log.Warn("unauthorized request: username not found or invalid password")
		return ctx.JSON(http.StatusUnauthorized, "{\"message\":\"unauthorized request: username not found or invalid password.\"}")
	}

	// Set custom claims
	claims := &goserver.JwtCustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "",
			Audience:  nil,
			Issuer:    "",
			Subject:   "",
			ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(time.Minute * time.Duration(s.JwtDuration))},
			IssuedAt:  &jwt.NumericDate{Time: time.Now()},
			NotBefore: nil,
		},
		Id:       999999,
		Name:     "Bill Whatever",
		Email:    "bill@whatever.com",
		Username: s.adminUser,
		IsAdmin:  true,
	}
	// create a uuid session id (a good place to store it in db if needed)
	sessionId, _ := uuid.NewUUID()

	// Create token with claims
	signer, _ := jwt.NewSignerHS(jwt.HS512, s.JwtSecret)
	builder := jwt.NewBuilder(signer)
	token, err := builder.Build(claims)
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("LoginUser(%s) succesfull login for user id (%d)", claims.Username, claims.Id)
	s.Log.Info(msg)
	return ctx.JSON(http.StatusOK, echo.Map{
		"token":   token.String(),
		"session": sessionId,
	})
}

func (s ServiceThing) restricted(ctx echo.Context) error {
	s.Log.Debug("++ entering restricted zone()")
	// get the current user from JWT TOKEN
	u := ctx.Get("jwtdata").(*jwt.Token)
	claims := goserver.JwtCustomClaims{}
	err := u.DecodeClaims(&claims)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	return ctx.JSON(http.StatusCreated, claims)
}

func (s ServiceThing) GetStatus(ctx echo.Context) error {
	s.Log.Debug("trace: entering GetStatus()")
	// get the current user from JWT TOKEN
	u := ctx.Get("jwtdata").(*jwt.Token)
	claims := goserver.JwtCustomClaims{}
	err := u.DecodeClaims(&claims)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	username := claims.Username
	idUser := claims.Id
	res, err := json.Marshal(claims)
	if err != nil {
		echo.NewHTTPError(http.StatusInternalServerError, "JWT User Data Could Not Be Marshaled To Json")
	}
	s.Log.Info("info: GetStatus(user:%s, id:%d)", username, idUser)
	return ctx.JSONBlob(http.StatusOK, res)
}

func (s ServiceThing) IsDBAlive() bool {
	dbVer, err := s.dbConn.GetVersion()
	if err != nil {
		return false
	}
	if len(dbVer) < 2 {
		return false
	}
	return true
}

func (s ServiceThing) checkReady(string) bool {
	// we decide what makes us ready, is a valid  connection to the database
	if !s.IsDBAlive() {
		return false
	}
	return true
}

func (s ServiceThing) checkHealthy(string) bool {
	// you decide what makes you ready, may be it is the connection to the database
	//if !IsDBAlive() {
	//	return false
	//}
	return true
}

func main() {
	prefix := fmt.Sprintf("%s ", version.APP)
	//l := log.New(os.Stdout, prefix, log.Ldate|log.Ltime|log.Lshortfile)
	l, err := golog.NewLogger("zap", golog.DebugLevel, prefix)
	l.Info("Starting %s v:%s  rev:%s  build: %s", version.APP, version.VERSION, version.REVISION, version.BuildStamp)
	l.Info("Repository url: https://%s", version.REPOSITORY)
	secret, err := config.GetJwtSecretFromEnv()
	if err != nil {
		l.Fatal("💥💥 ERROR: 'in NewGoHttpServer config.GetJwtSecretFromEnv() got error: %v'\n", err)
	}
	tokenDuration, err := config.GetJwtDurationFromEnv(60)
	if err != nil {
		l.Fatal("💥💥 ERROR: 'in NewGoHttpServer config.GetJwtDurationFromEnv() got error: %v'\n", err)
	}
	dbDsn, err := config.GetPgDbDsnUrlFromEnv(defaultDBIp, defaultDBPort,
		tools.ToSnakeCase(version.APP), version.AppSnake, defaultDBSslMode)
	if err != nil {
		l.Fatal("💥💥 error doing config.GetPgDbDsnUrlFromEnv. error: %v\n", err)
	}
	db, err := database.GetInstance("pgx", dbDsn, runtime.NumCPU(), l)
	if err != nil {
		l.Fatal("💥💥 error doing users.GetPgxConn(postgres, dbDsn  : %v\n", err)
	}
	defer db.Close()

	// checking database connection
	dbVersion, err := db.GetVersion()
	if err != nil {
		l.Fatal("💥💥 error doing dbConn.GetVersion() error: %v", err)
	}
	l.Info("connected to db version : %s", dbVersion)
	// checking metadata information
	metadataService := metadata.Service{
		Log: l,
		Db:  db,
	}

	err = metadataService.CreateMetadataTableIfNeeded()
	if err != nil {
		l.Fatal("💥💥 error doing metadataService.CreateMetadataTableIfNeeded  error: %v", err)
	}

	found, ver, err := metadataService.GetServiceVersion(version.APP)
	if err != nil {
		l.Fatal("💥💥 error doing metadataService.CreateMetadataTableIfNeeded  error: %v\n", err)
	}
	if found {
		l.Info("service %s was found in metadata with version: %s", version.APP, ver)
	} else {
		l.Info("service %s was not found in metadata", version.APP)
	}
	err = metadataService.SetServiceVersion(version.APP, version.VERSION)
	if err != nil {
		l.Fatal("💥💥 error doing metadataService.SetServiceVersion  error: %v\n", err)
	}

	// example of go-migrate db migration with embed files in go program
	// https://github.com/golang-migrate/migrate
	// https://github.com/golang-migrate/migrate/blob/master/database/postgres/TUTORIAL.md
	d, err := iofs.New(sqlMigrations, defaultSqlDbMigrationsPath)
	if err != nil {
		l.Fatal("💥💥 error doing iofs.New for db migrations  error: %v\n", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, strings.Replace(dbDsn, "postgres", "pgx", 1))
	if err != nil {
		l.Fatal("💥💥 error doing migrate.NewWithSourceInstance(iofs, dbURL:%s)  error: %v\n", dbDsn, err)
	}

	err = m.Up()
	if err != nil {
		//if err == m.
		if !errors.Is(err, migrate.ErrNoChange) {
			l.Fatal("💥💥 error doing migrate.Up error: %v\n", err)
		}
	}

	// set local admin user for test
	adminUsername := config.GetAdminUserFromFromEnv(defaultThingAdminUsername)
	adminPassword, err := config.GetAdminPasswordFromFromEnv()
	if err != nil {
		l.Fatal("💥💥 error GetAdminPasswordFromFromEnv unable to retrieve a valid admin password  error : %v'", err)
	}
	h := sha256.New()
	h.Write([]byte(adminPassword))
	adminPasswordHash := fmt.Sprintf("%x", h.Sum(nil))

	yourService := ServiceThing{
		Log:         l,
		dbConn:      db,
		JwtSecret:   []byte(secret),
		JwtDuration: tokenDuration,
		adminUser:   adminUsername,
		adminHash:   adminPasswordHash,
	}

	listenAddr, err := config.GetPortFromEnv(defaultPort)
	if err != nil {
		l.Fatal("💥💥 ERROR: 'calling GetPortFromEnv got error: %v'\n", err)
	}
	l.Info("Will start HTTP server listening on port %s", listenAddr)
	server := goserver.NewGoHttpServer(listenAddr, l, defaultWebRootDir, content, defaultSecuredApi)
	e := server.GetEcho()
	// https://echo.labstack.com/docs/middleware/prometheus
	mwConfig := echoprometheus.MiddlewareConfig{
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Path(), "/health")
		}, // does not gather metrics on routes starting with `/health`
		Subsystem: version.APP,
	}
	e.Use(echoprometheus.NewMiddlewareWithConfig(mwConfig)) // adds middleware to gather metrics
	e.GET("/metrics", echoprometheus.NewHandler())          // adds route to serve gathered metrics
	e.GET("/readiness", server.GetReadinessHandler(yourService.checkReady, "Connection to DB"))
	e.GET("/health", server.GetHealthHandler(yourService.checkHealthy, "Connection to DB"))
	//TODO  Find a way to allow Login route to be available only in dev environment
	e.POST("/login", yourService.login)
	r := server.GetRestrictedGroup()
	r.GET("/secret", yourService.restricted)
	r.GET("/status", yourService.GetStatus)

	thingStore, err := thing.GetStorageInstance("pgx", db, l)
	if err != nil {
		l.Fatal("💥💥 ERROR: 'calling things.GetStorageInstance got error: %v'\n", err)
	}
	// now with restricted group reference you can register your secured handlers defined in OpenApi things.yaml
	objService := thing.Service{
		Log:              l,
		DbConn:           db,
		Store:            thingStore,
		JwtSecret:        []byte(secret),
		JwtDuration:      tokenDuration,
		ListDefaultLimit: 10,
	}
	thing.RegisterHandlers(r, &objService)

	err = server.StartServer()
	if err != nil {
		l.Fatal("💥💥 ERROR: 'calling echo.Start(%s) got error: %v'\n", listenAddr, err)
	}

}
