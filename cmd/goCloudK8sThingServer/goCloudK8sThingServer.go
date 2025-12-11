package main

import (
	"embed"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goHttpEcho"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/metadata"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/tools"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-thing/pkg/thing"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-thing/pkg/version"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	defaultPort                = 9090
	defaultLogName             = "stderr"
	defaultDBPort              = 5432
	defaultDBIp                = "127.0.0.1"
	defaultDBSslMode           = "prefer"
	defaultJwtStatusUrl        = "/status"
	defaultJwtCookieName       = "goJWT_token"
	defaultAppInfoUrl          = "/goAppInfo"
	defaultWebRootDir          = "goCloudK8sThingFront/dist/"
	defaultSqlDbMigrationsPath = "db/migrations"
	defaultSecuredApi          = "/goapi/v1"
	defaultAdminUser           = "goadmin"
	defaultAdminEmail          = "goadmin@yourdomain.org"
	defaultAdminId             = 960901
	charsetUTF8                = "charset=UTF-8"
	MIMEAppJSON                = "application/json"
	MIMEHtml                   = "text/html"
	MIMEHtmlCharsetUTF8        = MIMEHtml + "; " + charsetUTF8
	MIMEAppJSONCharsetUTF8     = MIMEAppJSON + "; " + charsetUTF8
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

// UserLogin defines model for UserLogin.
type UserLogin struct {
	PasswordHash string `json:"password_hash"`
	Username     string `json:"username"`
}

type Service struct {
	Logger        golog.MyLogger
	dbConn        database.DB
	server        *goHttpEcho.Server
	jwtCookieName string
}

// login is just a trivial example to test this server
// you should use the jwt token returned from LoginUser  in github.com/lao-tseu-is-alive/go-cloud-k8s-user-group'
// and share the same secret with the above component
func (s *Service) login(ctx echo.Context) error {
	goHttpEcho.TraceHttpRequest("login", ctx.Request(), s.Logger)
	uLogin := new(UserLogin)
	login := ctx.FormValue("login")
	passwordHash := ctx.FormValue("hashed")
	s.Logger.Debug("login: %s, hash: %s ", login, passwordHash)
	// maybe it was not a form but a fetch data post
	if len(strings.Trim(login, " ")) < 1 {
		if err := ctx.Bind(uLogin); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid user login or json format in request body")
		}
	} else {
		uLogin.Username = login
		uLogin.PasswordHash = passwordHash
	}
	s.Logger.Debug("About to check username: %s , password: %s", uLogin.Username, uLogin.PasswordHash)

	if s.server.Authenticator.AuthenticateUser(uLogin.Username, uLogin.PasswordHash) {
		userInfo, err := s.server.Authenticator.GetUserInfoFromLogin(login)
		if err != nil {
			myErrMsg := fmt.Sprintf("Error getting user info from login: %v", err)
			s.Logger.Error(myErrMsg)
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"jwtStatus": myErrMsg, "token": ""})
		}
		token, err := s.server.JwtCheck.GetTokenFromUserInfo(userInfo)
		if err != nil {
			myErrMsg := fmt.Sprintf("Error getting jwt token from user info: %v", err)
			s.Logger.Error(myErrMsg)
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"jwtStatus": myErrMsg, "token": ""})
		}
		// Prepare the response
		response := map[string]string{
			"jwtStatus": "success",
			"token":     token.String(),
		}
		s.Logger.Info("LoginUser(%s) successful login", login)
		return ctx.JSON(http.StatusOK, response)
	} else {
		myErrMsg := "username not found or password invalid"
		s.Logger.Warn(myErrMsg)
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"jwtStatus": myErrMsg, "token": ""})
	}
}

func (s *Service) GetStatus(ctx echo.Context) error {
	goHttpEcho.TraceHttpRequest("GetStatus", ctx.Request(), s.Logger)
	// get the current user from JWT TOKEN
	claims := s.server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	s.Logger.Info("in restricted : currentUserId: %d", currentUserId)
	// you can check if the user is not active anymore and RETURN 401 Unauthorized
	//if !s.Store.IsUserActive(currentUserId) {
	//	return echo.NewHTTPError(http.StatusUnauthorized, "current calling user is not active anymore")
	//}
	return ctx.JSON(http.StatusOK, claims)
}

func (s *Service) IsDBAlive() bool {
	dbVer, err := s.dbConn.GetVersion()
	if err != nil {
		return false
	}
	if len(dbVer) < 2 {
		return false
	}
	return true
}

func (s *Service) checkReady(string) bool {
	// we decide what makes us ready, is a valid  connection to the database
	if !s.IsDBAlive() {
		return false
	}
	return true
}

func checkHealthy(string) bool {
	// you decide what makes you ready, may be it is the connection to the database
	//if !IsDBAlive() {
	//	return false
	//}
	return true
}

func initMetadataOrFail(db database.DB, l golog.MyLogger) {
	// checking metadata information
	metadataService := metadata.Service{Log: l, Db: db}
	metadataService.CreateMetadataTableOrFail()
	found, ver := metadataService.GetServiceVersionOrFail(version.APP)
	if found {
		l.Info("service %s was found in metadata with version: %s", version.APP, ver)
	} else {
		l.Info("service %s was not found in metadata", version.APP)
	}
	metadataService.SetServiceVersionOrFail(version.APP, version.VERSION)
}

func runMigrationsOrFail(dbDsn string, l golog.MyLogger) {
	// begin section go-migrate db migration with embed files in go program
	// https://github.com/golang-migrate/migrate
	d, err := iofs.New(sqlMigrations, defaultSqlDbMigrationsPath)
	if err != nil {
		l.Fatal("💥💥 error doing iofs.New for db migrations  error: %v\n", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, strings.Replace(dbDsn, "postgres", "pgx5", 1))
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
	// end section go-migrate db migration with embed files in go program
}

func main() {
	l, err := golog.NewLogger(
		"simple", // can be "zap"
		config.GetLogWriterFromEnvOrPanic(defaultLogName),
		config.GetLogLevelFromEnvOrPanic(golog.InfoLevel),
		version.APP,
	)
	if err != nil {
		panic(fmt.Sprintf("💥💥 error log.NewLogger error: %v'\n", err))
	}
	l.Info("🚀🚀 Starting App:'%s', ver:%s, from: %s", version.APP, version.VERSION, version.REPOSITORY)

	dbDsn := config.GetPgDbDsnUrlFromEnvOrPanic(defaultDBIp, defaultDBPort, tools.ToSnakeCase(version.APP), version.AppSnake, defaultDBSslMode)
	db, err := database.GetInstance("pgx", dbDsn, runtime.NumCPU(), l)
	if err != nil {
		l.Fatal("💥💥 error doing database.GetInstance(pgx ...) error: %v", err)
	}
	defer db.Close()

	dbVersion, err := db.GetVersion()
	if err != nil {
		l.Fatal("💥💥 error doing dbConn.GetVersion() error: %v", err)
	}
	l.Info("connected to db version : %s", dbVersion)

	initMetadataOrFail(db, l)
	runMigrationsOrFail(dbDsn, l)

	// Get the ENV JWT_AUTH_URL value
	jwtAuthUrl := config.GetJwtAuthUrlFromEnvOrPanic()
	jwtStatusUrl := config.GetJwtStatusUrlFromEnv(defaultJwtStatusUrl)

	myVersionReader := goHttpEcho.NewSimpleVersionReader(
		version.APP,
		version.VERSION,
		version.REPOSITORY,
		version.REVISION,
		version.BuildStamp,
		jwtAuthUrl,
		jwtStatusUrl,
	)
	// Create a new JWT checker
	myJwt := goHttpEcho.NewJwtChecker(
		config.GetJwtSecretFromEnvOrPanic(),
		config.GetJwtIssuerFromEnvOrPanic(),
		version.APP,
		config.GetJwtContextKeyFromEnvOrPanic(),
		config.GetJwtDurationFromEnvOrPanic(60),
		l)
	// Create a new Authenticator with a simple admin user
	myAuthenticator := goHttpEcho.NewSimpleAdminAuthenticator(&goHttpEcho.UserInfo{
		UserId:     config.GetAdminIdFromEnvOrPanic(defaultAdminId),
		ExternalId: config.GetAdminExternalIdFromEnvOrPanic(9999999),
		Name:       "NewSimpleAdminAuthenticator_Admin",
		Email:      config.GetAdminEmailFromEnvOrPanic(defaultAdminEmail),
		Login:      config.GetAdminUserFromEnvOrPanic(defaultAdminUser),
		IsAdmin:    false,
	},
		config.GetAdminPasswordFromEnvOrPanic(),
		myJwt)

	server := goHttpEcho.CreateNewServerFromEnvOrFail(
		defaultPort,
		"0.0.0.0", // defaultServerIp,
		&goHttpEcho.Config{
			ListenAddress: "",
			Authenticator: myAuthenticator,
			JwtCheck:      myJwt,
			VersionReader: myVersionReader,
			Logger:        l,
			WebRootDir:    defaultWebRootDir,
			Content:       content,
			RestrictedUrl: defaultSecuredApi,
		},
	)

	cookieNameForJWT := config.GetJwtCookieNameFromEnv(defaultJwtCookieName)
	yourService := Service{
		Logger:        l,
		dbConn:        db,
		server:        server,
		jwtCookieName: cookieNameForJWT,
	}

	e := server.GetEcho()
	//e.Use(goHttpEcho.CookieToHeaderMiddleware(yourService.jwtCookieName, l))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://golux.lausanne.ch", "http://localhost:3000"},
		AllowMethods:     []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
		AllowCredentials: true,
	}))

	// begin prometheus stuff to create a custom counter metric
	customCounter := prometheus.NewCounter( // create new counter metric. This is replacement for `prometheus.Metric` struct
		prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_custom_requests_total", version.APP),
			Help: "How many HTTP requests processed, partitioned by status code and HTTP method.",
		},
	)
	if err := prometheus.Register(customCounter); err != nil { // register your new counter metric with default metrics registry
		l.Fatal("💥💥 ERROR: 'calling prometheus.Register got error: %v'\n", err)
	}
	// https://echo.labstack.com/docs/middleware/prometheus
	mwConfig := echoprometheus.MiddlewareConfig{
		AfterNext: func(c echo.Context, err error) {
			customCounter.Inc() // use our custom metric in middleware. after every request increment the counter
		},
		// does not gather metrics on routes starting with `/health`
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Path(), "/health")
		},
		Subsystem: version.APP,
	}
	e.Use(echoprometheus.NewMiddlewareWithConfig(mwConfig)) // adds middleware to gather metrics
	// end prometheus stuff to create a custom counter metric

	e.GET("/metrics", echoprometheus.NewHandler()) // adds route to serve gathered metrics
	e.GET("/readiness", server.GetReadinessHandler(yourService.checkReady, "Connection to DB"))
	e.GET("/health", server.GetHealthHandler(checkHealthy, "Connection to DB"))
	e.GET(defaultAppInfoUrl, server.GetAppInfoHandler())
	// Find a way to allow Login route to be available only in dev environment
	e.POST(jwtAuthUrl, yourService.login)
	// Call the DevRoutes function conditionally
	// This line will only compile if the 'dev' build tag is active.
	// Conditional compilation of dev routes

	if IsDevBuild {
		l.Info("Attempting to register dev routes...")
		DevRoutes(e, &yourService, jwtAuthUrl)
	}
	r := server.GetRestrictedGroup()
	r.GET(jwtStatusUrl, yourService.GetStatus)

	thingStore := thing.GetStorageInstanceOrPanic("pgx", db, l)

	// now with restricted group reference you can register your secured handlers defined in OpenApi things.yaml
	thingService := thing.Service{
		Log:              l,
		DbConn:           db,
		Store:            thingStore,
		Server:           server,
		ListDefaultLimit: 50,
	}

	thing.RegisterHandlers(r, &thingService) // register all openapi declared routes

	err = server.StartServer()
	if err != nil {
		l.Fatal("💥💥 ERROR: 'calling echo.StartServer() got error: %v'\n", err)
	}

}
