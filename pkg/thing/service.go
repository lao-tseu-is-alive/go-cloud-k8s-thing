package thing

import (
	"fmt"
	"github.com/cristalhq/jwt/v4"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goserver"
	"net/http"
	"strings"
)

type Service struct {
	Log         golog.MyLogger
	dbConn      database.DB
	Store       Storage
	JwtSecret   []byte
	JwtDuration int
}

// List returns a list of things in the store and return it as json
// to test it with curl you can try :
// curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $token" 'http://localhost:9090/goapi/v1/thing' |jq
func (s Service) List(ctx echo.Context, params ListParams) error {
	s.Log.Info("trace: entering Thing List() params:%+v", params)
	// get the current user from JWT TOKEN
	u := ctx.Get("jwtdata").(*jwt.Token)
	claims := goserver.JwtCustomClaims{}
	err := u.DecodeClaims(&claims)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	limit := 10
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	offset := 0
	if params.Offset != nil {
		offset = int(*params.Offset)
	}
	list, err := s.Store.List(offset, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("there was a problem when calling store.List :%v", err))
	}
	return ctx.JSON(http.StatusOK, list)
}

func (s Service) Create(ctx echo.Context) error {
	s.Log.Debug("trace: entering Create()")
	// get the current user from JWT TOKEN
	u := ctx.Get("jwtdata").(*jwt.Token)
	claims := goserver.JwtCustomClaims{}
	err := u.DecodeClaims(&claims)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	// IF USER IS NOT ADMIN RETURN 401 Unauthorized
	currentUserId := claims.Id
	/* TODO implement ACL & RBAC handling
	if !s.Store.IsUserAllowedToCreate(currentUserId, typeThing) {
		return echo.NewHTTPError(http.StatusUnauthorized, "current user has no create role privilege")
	}
	*/
	newThing := &Thing{
		CreateBy: int32(currentUserId),
	}
	if err := ctx.Bind(newThing); err != nil {
		return ctx.JSON(http.StatusBadRequest, fmt.Sprintf("Create has invalid format [%v]", err))
	}
	if len(strings.Trim(newThing.Name, " ")) < 1 {
		return ctx.JSON(http.StatusBadRequest, fmt.Sprint("Create name cannot be empty or spaces only"))
	}
	if len(newThing.Name) < 5 {
		return ctx.JSON(http.StatusBadRequest, fmt.Sprintf("Create name minLength is 5 not (%d)", len(newThing.Name)))
	}
	s.Log.Info("# Create() newThing : %#v\n", newThing)
	thingCreated, err := s.Store.Create(*newThing)
	if err != nil {
		msg := fmt.Sprintf("CreateUser had an error saving user:%#v, err:%#v", *newThing, err)
		s.Log.Info(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	s.Log.Info("# Create() Thing %#v\n", thingCreated)
	return ctx.JSON(http.StatusCreated, thingCreated)
}

func (s Service) Delete(ctx echo.Context, thingId uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) Get(ctx echo.Context, thingId uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) Update(ctx echo.Context, thingId uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) ListByType(ctx echo.Context, typeId int32, params ListByTypeParams) error {
	//TODO implement me
	panic("implement me")
}
func (s Service) ListByExternalId(ctx echo.Context, externalId int32, params ListByExternalIdParams) error {
	//TODO implement me
	panic("implement me")
}
func (s Service) TypeThingList(ctx echo.Context, params TypeThingListParams) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) TypeThingCreate(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) TypeThingDelete(ctx echo.Context, typeThingId int32) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) TypeThingGet(ctx echo.Context, typeThingId int32) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) TypeThingUpdate(ctx echo.Context, typeThingId int32) error {
	//TODO implement me
	panic("implement me")
}
